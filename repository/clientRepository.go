package repository

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/hoffie/larasync/helpers"
	"github.com/hoffie/larasync/helpers/atomic"
	"github.com/hoffie/larasync/helpers/crypto"
	"github.com/hoffie/larasync/helpers/path"
	"github.com/hoffie/larasync/repository/chunker"
	"github.com/hoffie/larasync/repository/nib"
	"github.com/hoffie/larasync/repository/tracker"
)

// ClientRepository is a Repository from a client-side view; it has all the keys
// and a work dir (comapred to the base Repository)
type ClientRepository struct {
	*Repository
	stateConfig *StateConfig
	nibTracker  tracker.NIBTracker
}

// NewClient returns a new ClientRepository instance
func NewClient(path string) *ClientRepository {
	repo := New(path)
	return &ClientRepository{
		Repository: repo,
	}
}

// NIBTracker returns the
func (r *ClientRepository) NIBTracker() (tracker.NIBTracker, error) {
	if r.nibTracker == nil {
		repo := r.Repository
		tracker, err := tracker.NewDatabaseNIBTracker(
			filepath.Join(repo.GetManagementDir(), "nib_tracker.db"),
			repo.Path,
		)
		if err != nil {
			return nil, err
		}
		r.nibTracker = tracker
	}
	return r.nibTracker, nil
}

// StateConfig returns this repository's state config; it is currently used
// in client repositories only and stores things like the default server.
func (r *ClientRepository) StateConfig() (*StateConfig, error) {
	if r.stateConfig != nil {
		return r.stateConfig, nil
	}
	path := r.subPathFor(stateConfigFileName)
	r.stateConfig = NewStateConfig(path)
	err := r.stateConfig.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return r.stateConfig, nil
}

// writeFileToChunks takes a file path and saves its contents to the
// storage in encrypted form with a content-addressing id.
func (r *ClientRepository) writeFileToChunks(path string) ([]string, error) {
	return r.splitFileToChunks(path, r.writeCryptoContainerObject)
}

// getFileChunkIDs analyzes the given file and returns its content ids.
// This function does not write anything to disk.
func (r *ClientRepository) getFileChunkIDs(path string) ([]string, error) {
	return r.splitFileToChunks(path, func(string, []byte) error { return nil })
}

// splitFileToChunks takes a file path and splits its contents into chunks
// identified by their content ids.
func (r *ClientRepository) splitFileToChunks(path string, handler func(string, []byte) error) ([]string, error) {
	chunker, err := chunker.New(path, chunkSize)
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

// fileToChunkIds returnes te current chunk hashes for the given path.
func (r *ClientRepository) fileToChunkIds(path string) ([]string, error) {
	chunker, err := chunker.New(path, chunkSize)
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
	}
	return ids, nil
}

// GetSigningPrivateKey exposes the signing private key as it is required
// in foreign packages such as api.
func (r *ClientRepository) GetSigningPrivateKey() ([PrivateKeySize]byte, error) {
	return r.keys.SigningPrivateKey()
}

// CreateKeys handles creation of all required cryptographic keys.
func (r *ClientRepository) CreateKeys() error {
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
func (r *ClientRepository) encryptWithRandomKey(data []byte) ([]byte, error) {
	encryptionKey, err := r.keys.EncryptionKey()
	if err != nil {
		return nil, err
	}
	cryptoBox := crypto.NewBox(encryptionKey)
	return cryptoBox.EncryptWithRandomKey(data)
}

// hashChunk takes a chunk of data and constructs its content-addressing
// hash.
func (r *ClientRepository) hashChunk(chunk []byte) (string, error) {
	key, err := r.keys.HashingKey()
	if err != nil {
		return "", err
	}
	hasher := crypto.NewHasher(key)
	return hasher.StringHash(chunk), nil
}

// writeCryptoContainerObject takes a piece of raw data and
// writes it to the object store in encrypted form.
func (r *ClientRepository) writeCryptoContainerObject(id string, data []byte) error {
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
func (r *ClientRepository) readEncryptedObject(id string) ([]byte, error) {
	reader, err := r.objectStorage.Get(id)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	encryptedContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return r.decryptContent(encryptedContent)
}

// decryptContent is the counter-part of encryptWithRandomKey, i.e.
// it returns the plain text again.
func (r *ClientRepository) decryptContent(enc []byte) ([]byte, error) {
	encryptionKey, err := r.keys.EncryptionKey()
	if err != nil {
		return nil, err
	}
	cryptoBox := crypto.NewBox(encryptionKey)
	return cryptoBox.DecryptContent(enc)
}

// writeMetadata writes the metadata object for the given path
// to disk and returns its id.
func (r *ClientRepository) writeMetadata(absPath string) (string, error) {
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

// pathToNIBID returns the NIB ID for the given relative path
func (r *ClientRepository) pathToNIBID(relPath string) (string, error) {
	return r.hashChunk([]byte(relPath))
}

// metadataByID returns the metadata object identified by the given object id.
func (r *ClientRepository) metadataByID(id string) (*Metadata, error) {
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
func (r *ClientRepository) CheckoutPath(absPath string) error {
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
func (r *ClientRepository) CheckoutAllPaths() error {
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
func (r *ClientRepository) pathHasConflictingChanges(nib *nib.NIB, absPath string) (bool, error) {
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
func (r *ClientRepository) checkoutNIB(nib *nib.NIB) error {
	rev, err := nib.LatestRevision()
	if err != nil {
		return err
	}

	return r.checkoutRevision(nib, rev)
}

// checkoutRevision checks out the provided Revision into the working directory.
func (r *ClientRepository) checkoutRevision(nib *nib.NIB, rev *nib.Revision) error {
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
	err = nil

	if len(rev.ContentIDs) > 0 {
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
			return ErrWorkDirConflict
		}
	} else if _, errExistCheck := os.Stat(absPath); errExistCheck == nil {
		err = os.Remove(absPath)
	}

	return err
}

// AddItem adds a new file or directory to the repository.
func (r *ClientRepository) AddItem(absPath string) error {
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
	return r.addFile(absPath)
}

// addFile adds the given file from the working directory
// to the repository
func (r *ClientRepository) addFile(absPath string) error {
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

	n := &nib.NIB{ID: nibID}
	if nibStore.Exists(nibID) {
		n, err = nibStore.Get(nibID)
		if err != nil {
			return err
		}
	}

	rev := &nib.Revision{}
	rev.MetadataID = metadataID
	rev.ContentIDs = contentIDs
	rev.UTCTimestamp = time.Now().UTC().Unix()
	//FIXME: deviceID etc.
	latestRev, err := n.LatestRevision()
	if err != nil && err != nib.ErrNoRevision {
		return err
	}
	if err == nib.ErrNoRevision || !latestRev.HasSameContent(rev) {
		n.AppendRevision(rev)
	}
	err = r.notifyNIBTracker(nibID, relPath)
	if err != nil {
		return err
	}

	return nibStore.Add(n)
}

// notifyNIBTracker adds the passed relative path to the NIBTracker of
// this client repository.
func (r *ClientRepository) notifyNIBTracker(nibID string, relPath string) error {
	tracker, err := r.NIBTracker()
	if err != nil {
		return err
	}
	return tracker.Add(relPath, nibID)
}

// addDirectory walks the given directory and calls AddItem on each entry
func (r *ClientRepository) addDirectory(absPath string) error {
	files, err := ioutil.ReadDir(absPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(absPath, file.Name())
		err = r.AddItem(path)
		if err == ErrRefusingWorkOnDotLara {
			continue
		} else if err != nil {
			return err
		}
	}
	return nil
}

// DeleteItem removes the given item with the passed absolute path.
func (r *ClientRepository) DeleteItem(absPath string) error {
	relPath, err := r.getRepoRelativePath(absPath)
	if err != nil {
		return err
	}

	nibID, err := r.pathToNIBID(relPath)
	if err != nil {
		return err
	}

	nib, err := r.nibStore.Get(nibID)
	if os.IsNotExist(err) {
		return r.deleteDirectory(absPath)
	} else if err != nil {
		return err
	}
	rev, err := nib.LatestRevision()
	if err != nil {
		return err
	}

	if !rev.IsDeletion() {
		err = r.deleteFile(absPath)
	} else {
		err = r.deleteDirectory(absPath)
	}
	return err
}

// deleteDirectory checks the ClientRepositories path lookup and removes all
// files and directories in the given path.
func (r *ClientRepository) deleteDirectory(absPath string) error {
	relPath, err := r.getRepoRelativePath(absPath)
	if err != nil {
		return err
	}
	paths, err := r.nibTracker.SearchPrefix(relPath)
	if err != nil {
		return err
	}

	for _, path := range paths {
		err = r.DeleteItem(path.AbsPath())
		if err != nil {
			return err
		}
	}

	return path.CleanUpEmptyDirs(absPath)
}

// deleteFile removes the specific file from the repository. Returns an error
// if the file does not exist in the repository.
func (r *ClientRepository) deleteFile(absPath string) error {
	relPath, err := r.getRepoRelativePath(absPath)
	if err != nil {
		return err
	}

	nibID, err := r.pathToNIBID(relPath)
	if err != nil {
		return err
	}

	nibItem, err := r.nibStore.Get(nibID)
	if err != nil {
		return err
	}

	latestRevision, err := nibItem.LatestRevision()
	if err != nil && err != nib.ErrNoRevision {
		return err
	}

	deleteFileIfExisting := func() error {
		if r.revisionIsFile(absPath, latestRevision) && !latestRevision.IsDeletion() {
			os.Remove(absPath)
		}

		stat, fileErr := os.Stat(absPath)
		if fileErr != nil {
			return nil
		}

		if !stat.IsDir() && latestRevision != nil {
			ids, err := r.fileToChunkIds(absPath)
			if err != nil {
				return err
			}
			if helpers.StringsEqual(ids, latestRevision.ContentIDs) {
				return os.Remove(absPath)
			}
		}
		return nil
	}

	if err == nil && latestRevision != nil {
		if latestRevision.IsDeletion() {
			return deleteFileIfExisting()
		}
		deleteRevision := latestRevision.Clone()
		deleteRevision.ContentIDs = []string{}
		nibItem.AppendRevision(deleteRevision)
		err = r.nibStore.Add(nibItem)
		if err != nil {
			return err
		}
	}

	return deleteFileIfExisting()
}

// revisionIsFile returns if the given revision is represented by the passed revision.
func (r *ClientRepository) revisionIsFile(absPath string, rev *nib.Revision) bool {
	stat, err := os.Stat(absPath)
	if rev == nil || (err == nil && stat.IsDir()) {
		return false
	}

	if os.IsNotExist(err) {
		if rev.IsDeletion() {
			return true
		}
		return false
	}

	ids, err := r.fileToChunkIds(absPath)
	if err != nil {
		return false
	}

	return helpers.StringsEqual(ids, rev.ContentIDs)
}

// SetAuthorization adds a authorization with the given publicKey and encrypts it with the
// passed encryptionKey to this repository.
func (r *ClientRepository) SetAuthorization(
	publicKey [PublicKeySize]byte,
	encKey [EncryptionKeySize]byte,
	authorization *Authorization,
) error {
	return r.authorizationManager.Set(publicKey, encKey, authorization)
}

// NewAuthorization returns the currently valid Authorization object
// for this repository. If the privateKeys necessary for this are not
// stored in the keyStore an error is returned.
func (r *ClientRepository) NewAuthorization() (*Authorization, error) {
	encryptionKey, err := r.keys.EncryptionKey()
	if err != nil {
		return nil, errors.New("Could not load encryption key.")
	}

	hashingKey, err := r.keys.HashingKey()
	if err != nil {
		return nil, errors.New("Could not load hashing key.")
	}

	signatureKey, err := r.keys.SigningPrivateKey()
	if err != nil {
		return nil, errors.New("Could not load private signing key.")
	}

	auth := &Authorization{
		EncryptionKey: encryptionKey,
		HashingKey:    hashingKey,
		SigningKey:    signatureKey,
	}

	return auth, nil
}

// SerializedAuthorization returns a new, serialized authorization package.
func (r *ClientRepository) SerializedAuthorization(encryptionKey [EncryptionKeySize]byte) ([]byte, error) {
	auth, err := r.NewAuthorization()
	if err != nil {
		return nil, fmt.Errorf("authorization creation error (%s)", err)
	}

	authorizationBytes, err := r.SerializeAuthorization(encryptionKey, auth)
	if err != nil {
		return nil, fmt.Errorf("authorization encryption failure (%s)", err)
	}
	return authorizationBytes, nil
}

// TransactionsFrom returns all transactions which have been added since the given transactionID.
func (r *ClientRepository) TransactionsFrom(transactionID int64) ([]*Transaction, error) {
	return r.transactionManager.From(transactionID)
}
