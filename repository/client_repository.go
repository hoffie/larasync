package repository

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hoffie/larasync/helpers/atomic"
	"github.com/hoffie/larasync/helpers/crypto"
	"github.com/hoffie/larasync/helpers/path"
)

// ClientRepository is a Repository from a client-side view; it has all the keys
// and a work dir (comapred to the base Repository)
type ClientRepository struct {
	*Repository
	stateConfig *StateConfig
}

// NewClient returns a new ClientRepository instance
func NewClient(path string) *ClientRepository {
	return &ClientRepository{Repository: New(path)}
}

// StateConfig returns this repository's state config; it is currently used
// in client repositories only and stores things like the default server.
func (r *ClientRepository) StateConfig() (*StateConfig, error) {
	if r.stateConfig != nil {
		return r.stateConfig, nil
	}
	path := r.subPathFor(stateConfigFileName)
	r.stateConfig = &StateConfig{Path: path}
	err := r.stateConfig.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return r.stateConfig, nil
}

// writeFileToChunks takes a file path and saves its contents to the
// storage in encrypted form with a content-addressing id.
func (r *Repository) writeFileToChunks(path string) ([]string, error) {
	return r.splitFileToChunks(path, r.writeCryptoContainerObject)
}

// getFileChunkIDs analyzes the given file and returns its content ids.
// This function does not write anything to disk.
func (r *Repository) getFileChunkIDs(path string) ([]string, error) {
	return r.splitFileToChunks(path, func(string, []byte) error { return nil })
}

// splitFileToChunks takes a file path and splits its contents into chunks
// identified by their content ids.
func (r *Repository) splitFileToChunks(path string, handler func(string, []byte) error) ([]string, error) {
	chunker, err := NewChunker(path, chunkSize)
	if err != nil {
		return nil, err
	}
	defer chunker.Close()
	var ids []string
	for chunker.HasNext() {
		chunk, err := chunker.Next()
		if err != nil {
			return nil, err
		}

		// hash for content-addressing
		hexHash, err := r.hashChunk(chunk)
		if err != nil {
			return nil, err
		}

		ids = append(ids, hexHash)

		err = handler(hexHash, chunk)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// GetSigningPrivateKey exposes the signing private key as it is required
// in foreign packages such as api.
func (r *Repository) GetSigningPrivateKey() ([PrivateKeySize]byte, error) {
	return r.keys.SigningPrivateKey()
}

// CreateKeys handles creation of all required cryptographic keys.
func (r *Repository) CreateKeys() error {
	err := r.keys.CreateEncryptionKey()
	if err != nil {
		return err
	}

	err = r.keys.CreateSigningKey()
	if err != nil {
		return err
	}

	err = r.keys.CreateHashingKey()
	if err != nil {
		return err
	}

	return nil
}

// encryptWithRandomKey takes a piece of data, encrypts it with a random
// key and returns the result, prefixed by the random key encrypted by
// the repository encryption key.
func (r *Repository) encryptWithRandomKey(data []byte) ([]byte, error) {
	encryptionKey, err := r.keys.EncryptionKey()
	if err != nil {
		return nil, err
	}
	cryptoBox := crypto.NewBox(encryptionKey)
	return cryptoBox.EncryptWithRandomKey(data)
}

// hashChunk takes a chunk of data and constructs its content-addressing
// hash.
func (r *Repository) hashChunk(chunk []byte) (string, error) {
	key, err := r.keys.HashingKey()
	if err != nil {
		return "", err
	}
	hasher := crypto.NewHasher(key)
	return hasher.StringHash(chunk), nil
}

// writeCryptoContainerObject takes a piece of raw data and
// writes it to the object store in encrypted form.
func (r *Repository) writeCryptoContainerObject(id string, data []byte) error {
	// PERFORMANCE: avoid re-writing pre-existing metadata files by checking for
	// existance first.
	var enc []byte
	enc, err := r.encryptWithRandomKey(data)
	if err != nil {
		return err
	}

	err = r.AddObject(id, bytes.NewReader(enc))
	if err != nil {
		return err
	}

	return nil
}

// readEncryptedObject reads the object with the given id and returns its
// authenticated, unencrypted content.
func (r *Repository) readEncryptedObject(id string) ([]byte, error) {
	reader, err := r.objectStorage.Get(id)
	if err != nil {
		return nil, err
	}
	encryptedContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return r.decryptContent(encryptedContent)
}

// decryptContent is the counter-part of encryptWithRandomKey, i.e.
// it returns the plain text again.
func (r *Repository) decryptContent(enc []byte) ([]byte, error) {
	encryptionKey, err := r.keys.EncryptionKey()
	if err != nil {
		return nil, err
	}
	cryptoBox := crypto.NewBox(encryptionKey)
	return cryptoBox.DecryptContent(enc)
}

// writeMetadata writes the metadata object for the given path
// to disk and returns its id.
func (r *Repository) writeMetadata(absPath string) (string, error) {
	relPath, err := r.getRepoRelativePath(absPath)
	if err != nil {
		return "", err
	}
	m := Metadata{
		RepoRelativePath: relPath,
		Type:             MetadataTypeFile,
	}
	raw := &bytes.Buffer{}
	_, err = m.WriteTo(raw)
	if err != nil {
		return "", err
	}

	rawBytes := raw.Bytes()

	hexHash, err := r.hashChunk(rawBytes)
	if err != nil {
		return "", err
	}

	err = r.writeCryptoContainerObject(hexHash, rawBytes)
	if err != nil {
		return "", err
	}
	return hexHash, nil
}

// getFilesNIBUUID returns the NIB for the given relative path
func (r *Repository) pathToNIBID(relPath string) (string, error) {
	return r.hashChunk([]byte(relPath))
}

// metadataByID returns the metadata object identified by the given object id.
func (r *Repository) metadataByID(id string) (*Metadata, error) {
	rawMetadata, err := r.readEncryptedObject(id)
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{}
	_, err = metadata.ReadFrom(bytes.NewReader(rawMetadata))
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// CheckoutPath looks up the given path name in the internal repository state and
// writes the content from the repository state to the path in the working directory,
// possibly overwriting an existing version of the file.
func (r *Repository) CheckoutPath(absPath string) error {
	relPath, err := r.getRepoRelativePath(absPath)
	if err != nil {
		return err
	}

	id, err := r.pathToNIBID(relPath)
	if err != nil {
		return err
	}

	nibStore := r.nibStore

	// nibStore.Get also handles signature verification
	nib, err := nibStore.Get(id)
	if err != nil {
		return err
	}

	return r.checkoutNIB(nib)
}

// CheckoutAllPaths checks out all tracked paths.
func (r *Repository) CheckoutAllPaths() error {
	nibStore := r.nibStore
	nibs, err := nibStore.GetAll()
	if err != nil {
		return err
	}
	for nib := range nibs {
		err = r.checkoutNIB(nib)
		if err != nil {
			return err
		}
	}
	return nil
}

// pathHasConflictingChanges checks whether the item pointed to by absPath has any
// changes not resolvable to a revision in the given NIB.
func (r *Repository) pathHasConflictingChanges(nib *NIB, absPath string) (bool, error) {
	workdirContentIDs, err := r.getFileChunkIDs(absPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	_, err = nib.LatestRevisionWithContent(workdirContentIDs)
	return err != nil, nil
}

// checkoutNIB checks out the provided NIB's latest revision into the working directory.
func (r *Repository) checkoutNIB(nib *NIB) error {
	rev, err := nib.LatestRevision()
	if err != nil {
		return err
	}

	metadata, err := r.metadataByID(rev.MetadataID)
	if err != nil {
		return err
	}

	relPath := metadata.RepoRelativePath
	if relPath == "" {
		return errors.New("metadata lacks path")
	}
	absPath := filepath.Join(r.Path, relPath)

	targetDir := filepath.Dir(absPath)

	err = os.MkdirAll(targetDir, defaultDirPerms)
	if err != nil && !os.IsExist(err) {
		return err
	}

	writer, err := atomic.NewWriter(absPath, ".lara.checkout.", defaultFilePerms)
	defer writer.Close()
	if err != nil {
		writer.Abort()
		return err
	}

	for _, contentID := range rev.ContentIDs {
		content, err := r.readEncryptedObject(contentID)
		_, err = writer.Write(content)
		if err != nil {
			writer.Abort()
			return err
		}
	}

	hasChanges, err := r.pathHasConflictingChanges(nib, absPath)
	if err != nil {
		writer.Abort()
		return err
	}
	if hasChanges {
		writer.Abort()
		return errors.New("workdir conflict")
	}

	return nil
}

// AddItem adds a new file or directory to the repository.
func (r *Repository) AddItem(absPath string) error {
	stat, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	isBelow, err := path.IsBelow(absPath, filepath.Join(r.Path, managementDirName))
	if err != nil {
		return nil
	}
	if isBelow {
		return ErrRefusingWorkOnDotLara
	}
	if stat.IsDir() {
		return r.addDirectory(absPath)
	}
	metadataID, err := r.writeMetadata(absPath)
	if err != nil {
		return err
	}

	contentIDs, err := r.writeFileToChunks(absPath)
	if err != nil {
		return err
	}

	relPath, err := r.getRepoRelativePath(absPath)
	if err != nil {
		return err
	}
	nibID, err := r.pathToNIBID(relPath)
	if err != nil {
		return err
	}

	nibStore := r.nibStore

	nib := &NIB{ID: nibID}
	if nibStore.Exists(nibID) {
		nib, err = nibStore.Get(nibID)
		if err != nil {
			return err
		}
	}

	rev := &Revision{}
	rev.MetadataID = metadataID
	rev.ContentIDs = contentIDs
	if err != nil {
		return err
	}
	latestRev, err := nib.LatestRevision()
	if err != nil && err != ErrNoRevision {
		return err
	}
	if err == ErrNoRevision || !latestRev.HasSameContent(rev) {
		nib.AppendRevision(rev)
	}
	//FIXME: timestamp, deviceID etc.
	return nibStore.Add(nib)
}

// addDirectory walks the given directory and calls AddItem on each entry
func (r *Repository) addDirectory(absPath string) error {
	files, err := ioutil.ReadDir(absPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(absPath, file.Name())
		err = r.AddItem(path)
		if err == ErrRefusingWorkOnDotLara {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}
