package repository

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.google.com/p/go.crypto/nacl/secretbox"
)

const (
	// internal directory names
	managementDirName   = ".lara"
	objectsDirName      = "objects"
	nibsDirName         = "nibs"
	transactionsDirName = "transaction"

	// default permissions
	defaultFilePerms = 0600
	defaultDirPerms  = 0700

	// default chunk splitting size
	defaultChunkSize = 1 * 1024 * 1024

	// EncryptionKeySize represents the size of the key used for
	// encrypting.
	EncryptionKeySize = 32

	// HashingKeySize represents the size of the key used for
	// generating content hashes (HMAC).
	HashingKeySize = 32

	// secretbox nonceSize
	nonceSize = 24

	// pre-computed minimal length of ciphertext; anything less cannot be valid
	// and will be rejected before attempting any other operations.
	encryptedContentMinSize = 2*(nonceSize+secretbox.Overhead) + EncryptionKeySize
)

// Repository represents an on-disk repository and provides methods to
// access its sub-items.
type Repository struct {
	Path               string
	keys               *KeyStore
	objectStorage      ContentStorage
	nibStore           *NIBStore
	transactionManager *TransactionManager
	managementDirPath  string
	objectsPath        string
	nibsPath           string
}

// New returns a new repository instance with the given base path
func New(path string) *Repository {
	r := &Repository{Path: path}
	r.setupPaths()
	r.keys = NewKeyStore(r.managementDirPath)
	return r
}

// setupPaths initializes several attributes referring to internal repository paths
// such as encryption key paths.
func (r *Repository) setupPaths() {
	base := filepath.Join(r.Path, managementDirName)
	r.managementDirPath = base
	r.objectsPath = filepath.Join(base, objectsDirName)
	r.nibsPath = filepath.Join(base, nibsDirName)
}

// getObjectStorage returns the currently configured content storage backend
// for the repository.
func (r *Repository) getObjectStorage() (ContentStorage, error) {
	if r.objectStorage == nil {
		storage := FileContentStorage{
			StoragePath: filepath.Join(
				r.GetManagementDir(),
				objectsDirName)}
		err := storage.CreateDir()
		if err != nil {
			return nil, err
		}
		r.objectStorage = &storage
	}
	return r.objectStorage, nil
}

// getTransactionManager returns the currently configured
// transaction manager for the repository.
func (r *Repository) getTransactionManager() (*TransactionManager, error) {
	if r.transactionManager == nil {
		transactionStorage := FileContentStorage{
			StoragePath: filepath.Join(
				r.GetManagementDir(),
				transactionsDirName,
			)}
		err := transactionStorage.CreateDir()
		if err != nil {
			return nil, err
		}

		r.transactionManager = newTransactionManager(
			transactionStorage,
			r.GetManagementDir(),
		)
	}
	return r.transactionManager, nil
}

// getNIBStore returns the currently configured nib store
// for the repository.
func (r *Repository) getNIBStore() (*NIBStore, error) {
	if r.nibStore == nil {
		nibStorage := FileContentStorage{
			StoragePath: filepath.Join(
				r.GetManagementDir(),
				nibsDirName)}
		err := nibStorage.CreateDir()
		if err != nil {
			return nil, err
		}

		transactionManager, err := r.getTransactionManager()
		if err != nil {
			return nil, err
		}

		storage := ContentStorage(nibStorage)

		r.nibStore = newNIBStore(storage, r.keys, transactionManager)
	}
	return r.nibStore, nil
}

// CreateManagementDir ensures that this repository's management
// directory exists.
func (r *Repository) CreateManagementDir() error {
	err := os.Mkdir(r.GetManagementDir(), defaultDirPerms)
	if err != nil && !os.IsExist(err) {
		return err
	}
	path := r.nibsPath
	err = os.Mkdir(path, defaultDirPerms)
	if err != nil && !os.IsExist(err) {
		return err
	}
	_, err = r.getObjectStorage()
	if err != nil {
		return err
	}

	return nil
}

// GetManagementDir returns the path to the management directory.
func (r *Repository) GetManagementDir() string {
	return filepath.Join(r.Path, managementDirName)
}

// Create initially creates the repository directory structure.
func (r *Repository) Create() error {
	err := os.Mkdir(r.Path, defaultDirPerms)
	if err != nil {
		return err
	}
	err = r.CreateManagementDir()
	return err
}

// AddItem adds a new file or directory to the repository.
func (r *Repository) AddItem(absPath string) error {
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

	nibStore, err := r.getNIBStore()
	if err != nil {
		return err
	}

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

	nibStore, err := r.getNIBStore()
	if err != nil {
		return err
	}

	// nibStore.Get also handles signature verification
	nib, err := nibStore.Get(id)
	if err != nil {
		return err
	}

	rev, err := nib.LatestRevision()
	if err != nil {
		return err
	}

	rawMetadata, err := r.readEncryptedObject(rev.MetadataID)
	if err != nil {
		return err
	}

	metadata := &Metadata{}
	_, err = metadata.ReadFrom(bytes.NewReader(rawMetadata))
	if err != nil {
		return err
	}

	if metadata.RepoRelativePath != relPath {
		return errors.New("metadata name mismatch")
	}

	//FIXME make atomic (rename) / write to tempfile
	out, err := os.OpenFile(absPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, contentID := range rev.ContentIDs {
		content, err := r.readEncryptedObject(contentID)
		_, err = out.Write(content)
		if err != nil {
			return err
		}
	}
	return nil
}

// readEncryptedObject reads the object with the given id and returns its
// authenticated, unencrypted content.
func (r *Repository) readEncryptedObject(id string) ([]byte, error) {
	objectStorage, err := r.getObjectStorage()
	if err != nil {
		return nil, err
	}

	reader, err := objectStorage.Get(id)
	if err != nil {
		return nil, err
	}
	encryptedContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return r.decryptContent(encryptedContent)
}

// getFilesNIBUUID returns the NIB for the given relative path
func (r *Repository) pathToNIBID(relPath string) (string, error) {
	return r.hashChunk([]byte(relPath))
}

// AddObject adds an object into the storage with the given
// id and adds the data in the reader to it.
func (r *Repository) AddObject(objectID string, data io.Reader) error {
	storage, err := r.getObjectStorage()
	if err != nil {
		return err
	}
	return storage.Set(objectID, data)
}

// AddNIBContent adds NIBData to the repository after verifying it.
func (r *Repository) AddNIBContent(nibID string, nibReader io.Reader) error {
	nibStore, err := r.getNIBStore()
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(nibReader)
	if err != nil {
		return err
	}

	err = nibStore.VerifyContent(data)
	if err != nil {
		return err
	}

	return nibStore.AddContent(nibID, bytes.NewReader(data))
}

// GetNIB returns a NIB for the given ID in this repository.
func (r *Repository) GetNIB(id string) (*NIB, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}

	return store.Get(id)
}

// GetNIBReader returns the NIB with the given id in this repository.
func (r *Repository) GetNIBReader(id string) (io.ReadCloser, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}

	return store.getReader(id)
}

// GetNIBBytesFrom returns the signed byte structure for NIBs from the given
// transaction id
func (r *Repository) GetNIBBytesFrom(fromTransactionID int64) (<-chan []byte, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}

	return store.GetBytesFrom(fromTransactionID)
}

// GetNIBsFrom returns nibs added since the passed transaction ID.
func (r *Repository) GetNIBsFrom(fromTransactionID int64) (<-chan *NIB, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}
	return store.GetFrom(fromTransactionID)
}

// GetAllNIBBytes returns all NIBs signed byte representations in this repository.
func (r *Repository) GetAllNIBBytes() (<-chan []byte, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}

	return store.GetAllBytes()
}

// GetAllNibs returns all the nibs which are stored in this repository.
// Those will be returned with the oldest one first and the newest added
// last.
func (r *Repository) GetAllNibs() (<-chan *NIB, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}
	return store.GetAll()
}

// HasNIB checks if a NIB with the given ID exists in the repository.
func (r *Repository) HasNIB(id string) bool {
	store, err := r.getNIBStore()
	if err != nil {
		return false
	}
	return store.Exists(id)
}

// CurrentTransaction returns the currently newest Transaction for this
// repository.
func (r *Repository) CurrentTransaction() (*Transaction, error) {
	tm, err := r.getTransactionManager()
	if err != nil {
		return nil, err
	}
	return tm.CurrentTransaction()
}

// GetObjectData returns the data stored for the given objectID in this
// repository.
func (r *Repository) GetObjectData(objectID string) (io.Reader, error) {
	storage, err := r.getObjectStorage()
	if err != nil {
		return nil, err
	}
	return storage.Get(objectID)
}

// getRepoRelativePath turns the given path into a path relative to the
// repository root and returns it.
func (r *Repository) getRepoRelativePath(absPath string) (string, error) {
	if len(absPath) < len(r.Path)+1 {
		return "", errors.New("unable to resolve path: path too short")
	}
	rel := absPath[len(r.Path)+1:]
	return rel, nil
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
	//PERFORMANCE: avoid re-writing pre-existing metadata files by checking for
	// existance first.
	cid, err := r.writeContentAddressedCryptoContainer(raw.Bytes())
	if err != nil {
		return "", err
	}
	return cid, nil
}

// writeContentAddressedCryptoContainer takes a piece of raw data and
// streams it to disk in one content-addressed chunk while encrypting the
// data in the process.
func (r *Repository) writeContentAddressedCryptoContainer(data []byte) (string, error) {
	// hash for content-addressing
	hexHash, err := r.hashChunk(data)
	if err != nil {
		return "", err
	}

	var enc []byte
	enc, err = r.encryptWithRandomKey(data)
	if err != nil {
		return "", err
	}

	err = r.AddObject(hexHash, bytes.NewReader(enc))
	if err != nil {
		return "", err
	}

	return hexHash, nil
}

// encryptWithRandomKey takes a piece of data, encrypts it with a random
// key and returns the result, prefixed by the random key encrypted by
// the repository encryption key.
func (r *Repository) encryptWithRandomKey(data []byte) ([]byte, error) {
	// first generate and encrypt the per-file key and append it to
	// the result buffer:
	var nonce1 [nonceSize]byte
	_, err := rand.Read(nonce1[:])
	if err != nil {
		return nil, err
	}

	var fileKey [32]byte
	_, err = rand.Read(fileKey[:])
	if err != nil {
		return nil, err
	}
	repoKey, err := r.keys.EncryptionKey()
	if err != nil {
		return nil, err
	}
	out := nonce1[:]
	out = secretbox.Seal(out, fileKey[:], &nonce1, &repoKey)

	// then append the actual encrypted contents
	var nonce2 [nonceSize]byte
	_, err = rand.Read(nonce2[:])
	if err != nil {
		return nil, err
	}
	out = append(out, nonce2[:]...)
	out = secretbox.Seal(out, data, &nonce2, &fileKey)
	return out, nil
}

// decryptContent is the counter-part of encryptWithRandomKey, i.e.
// it returns the plain text again.
func (r *Repository) decryptContent(enc []byte) ([]byte, error) {
	if len(enc) < encryptedContentMinSize {
		return nil, errors.New("truncated ciphertext")
	}

	// first decrypt the file-specific key using the master key
	var nonce [nonceSize]byte
	readNonce := func() {
		copy(nonce[:], enc[:nonceSize])
		enc = enc[nonceSize:]
	}
	readNonce()
	repoKey, err := r.keys.EncryptionKey()
	if err != nil {
		return nil, err
	}
	l := EncryptionKeySize + secretbox.Overhead
	encryptedFileKey := enc[:l]
	enc = enc[l:]
	var fileKey []byte
	fileKey, success := secretbox.Open(fileKey, encryptedFileKey, &nonce, &repoKey)
	if !success {
		return nil, errors.New("file key decryption failed")
	}

	var arrFileKey [EncryptionKeySize]byte
	copy(arrFileKey[:], fileKey)

	readNonce()
	var content []byte
	content, success = secretbox.Open(content, enc, &nonce, &arrFileKey)
	if !success {
		return nil, errors.New("content decryption failed")
	}

	return content, nil
}

// hashChunk takes a chunk of data and constructs its content-addressing
// hash.
func (r *Repository) hashChunk(chunk []byte) (string, error) {
	key, err := r.keys.HashingKey()
	if err != nil {
		return "", err
	}
	hasher := hmac.New(sha512.New, key[:])
	hasher.Write(chunk)
	hash := hasher.Sum(nil)
	hexHash := hex.EncodeToString(hash)
	return hexHash, nil
}

// writeFileToChunks takes a file path and saves its contents to the
// storage in encrypted form with a content-addressing id.
func (r *Repository) writeFileToChunks(path string) ([]string, error) {
	chunker, err := NewChunker(path, defaultChunkSize)
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
		id, err := r.writeContentAddressedCryptoContainer(chunk)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetSigningPublicKey exposes the signing public key as it is required
// in foreign packages such as api.
func (r *Repository) GetSigningPublicKey() ([PublicKeySize]byte, error) {
	return r.keys.SigningPublicKey()
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
