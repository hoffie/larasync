package repository

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.google.com/p/go.crypto/nacl/secretbox"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
)

const (
	encryptionKeyFileName  = "encryption.key"
	hashingKeyFileName     = "hashing.key"
	signingPrivkeyFileName = "signing.priv"
	signingPubkeyFileName  = "signing.pub"
	managementDirName      = ".lara"
	objectsDirName         = "objects"
	nibsDirName            = "nibs"
	transactionsDirName    = "transaction"
	defaultFilePerms       = 0600
	defaultDirPerms        = 0700
	defaultChunkSize       = 1 * 1024 * 1024
	// EncryptionKeySize represents the size of the key used for
	// encrypting.
	EncryptionKeySize = 32
	// HashingKeySize represents the size of the key used for
	// generating content hashes (HMAC).
	HashingKeySize = 32
)

// ErrNoNIB is returned if no matching NIB can be found
// (see Repository.getFilesNIBUUID)
var ErrNoNIB = errors.New("no such NIB")

// Repository represents an on-disk repository and provides methods to
// access its sub-items.
type Repository struct {
	Path               string
	storage            ContentStorage
	nibStore           NIBStore
	encryptionKeyPath  string
	signingPrivkeyPath string
	signingPubkeyPath  string
	hashingKeyPath     string
	objectsPath        string
	nibsPath           string
}

// New returns a new repository instance with the given base path
func New(path string) *Repository {
	r := &Repository{Path: path}
	r.setupPaths()
	return r
}

// setupPaths initializes several attributes referring to internal repository paths
// such as encryption key paths.
func (r *Repository) setupPaths() {
	base := filepath.Join(r.Path, managementDirName)
	r.encryptionKeyPath = filepath.Join(base, encryptionKeyFileName)
	r.signingPrivkeyPath = filepath.Join(base, signingPrivkeyFileName)
	r.signingPubkeyPath = filepath.Join(base, signingPubkeyFileName)
	r.hashingKeyPath = filepath.Join(base, hashingKeyFileName)
	r.objectsPath = filepath.Join(base, objectsDirName)
	r.nibsPath = filepath.Join(base, nibsDirName)
}

// getStorage returns the currently configured content storage backend
// for the repository.
func (r *Repository) getStorage() (ContentStorage, error) {
	if r.storage == nil {
		storage := FileContentStorage{
			StoragePath: filepath.Join(
				r.GetManagementDir(),
				objectsDirName)}
		err := storage.CreateDir()
		if err != nil {
			return nil, err
		}
		r.storage = storage
	}
	return r.storage, nil
}

// getNIBStore returns the currently configured nib store backend
// for the repository.
func (r *Repository) getNIBStore() (NIBStore, error) {
	if r.nibStore == nil {
		nibStorage := FileContentStorage{
			StoragePath: filepath.Join(
				r.GetManagementDir(),
				nibsDirName)}
		err := nibStorage.CreateDir()
		if err != nil {
			return nil, err
		}

		storage := ContentStorage(nibStorage)

		transactionStorage := FileContentStorage{
			StoragePath: filepath.Join(
				r.GetManagementDir(),
				transactionsDirName,
			)}
		err = transactionStorage.CreateDir()
		if err != nil {
			return nil, err
		}

		transactionManager := newTransactionManager(transactionStorage)
		clientNIBStore := newClientNIBStore(&storage, r, transactionManager)

		nibStore := NIBStore(clientNIBStore)
		r.nibStore = nibStore
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
	_, err = r.getStorage()
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

// CreateEncryptionKey generates a random encryption key.
func (r *Repository) CreateEncryptionKey() error {
	key := make([]byte, EncryptionKeySize)
	var arrKey [EncryptionKeySize]byte
	_, err := rand.Read(key)
	if err != nil {
		return err
	}
	copy(arrKey[:], key)
	err = r.SetEncryptionKey(arrKey)
	return err
}

// CreateSigningKey generates a random signing key.
func (r *Repository) CreateSigningKey() error {
	key := make([]byte, PrivateKeySize)
	var arrKey [PrivateKeySize]byte
	_, err := rand.Read(key)
	if err != nil {
		return err
	}
	copy(arrKey[:], key)
	err = r.SetSigningPrivkey(arrKey)
	return err
}

// CreateHashingKey generates a random hashing key.
func (r *Repository) CreateHashingKey() error {
	key := make([]byte, HashingKeySize)
	var arrKey [HashingKeySize]byte
	_, err := rand.Read(key)
	if err != nil {
		return err
	}
	copy(arrKey[:], key)
	err = r.SetHashingKey(arrKey)
	return err
}

// SetEncryptionKey sets the repository encryption key
func (r *Repository) SetEncryptionKey(key [EncryptionKeySize]byte) error {
	return ioutil.WriteFile(r.encryptionKeyPath, key[:], defaultFilePerms)
}

// GetEncryptionKey returns the repository encryption key.
func (r *Repository) GetEncryptionKey() ([EncryptionKeySize]byte, error) {
	key, err := ioutil.ReadFile(r.encryptionKeyPath)
	if len(key) != EncryptionKeySize {
		return [EncryptionKeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [EncryptionKeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
}

// SetSigningPrivkey sets the repository signing private key
func (r *Repository) SetSigningPrivkey(key [PrivateKeySize]byte) error {
	return ioutil.WriteFile(r.signingPrivkeyPath, key[:], defaultFilePerms)
}

// GetSigningPrivkey returns the repository signing private key.
func (r *Repository) GetSigningPrivkey() ([PrivateKeySize]byte, error) {
	key, err := ioutil.ReadFile(r.signingPrivkeyPath)
	if len(key) != PrivateKeySize {
		return [PrivateKeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [PrivateKeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
}

// SetSigningPubkey sets the repository signing key's public key.
func (r *Repository) SetSigningPubkey(key []byte) error {
	return ioutil.WriteFile(r.signingPubkeyPath, key, defaultFilePerms)
}

// GetSigningPubkey returns the repository signing public key.
func (r *Repository) GetSigningPubkey() ([PubkeySize]byte, error) {
	privKey, err := r.GetSigningPrivkey()
	if err != nil {
		return r.getSigningPubkeyFromFile()
	}
	return edhelpers.GetPublicKeyFromPrivate(privKey), nil
}

// getSigningPubkeyFromFile returns the repository signing public key.
//
// It tries to retrieve the stored copy and is only called if the public key
// cannot be derived from the private key (i.e. if the private key is not
// available in this repository).
func (r *Repository) getSigningPubkeyFromFile() ([PubkeySize]byte, error) {
	key, err := ioutil.ReadFile(r.signingPubkeyPath)
	if len(key) != PubkeySize {
		return [PubkeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [PubkeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
}

// SetHashingKey sets the repository hashing key (content addressing)
func (r *Repository) SetHashingKey(key [HashingKeySize]byte) error {
	return ioutil.WriteFile(r.hashingKeyPath, key[:], defaultFilePerms)
}

// GetHashingKey returns the repository signing private key.
func (r *Repository) GetHashingKey() ([HashingKeySize]byte, error) {
	key, err := ioutil.ReadFile(r.hashingKeyPath)
	if len(key) != HashingKeySize {
		return [HashingKeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [HashingKeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
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
	uuid, err := r.getFilesNIBUUID(relPath)
	if err != nil && err != ErrNoNIB {
		return err
	}

	rev := &Revision{}
	rev.MetadataID = metadataID
	rev.ContentIDs = contentIDs
	nib := NIB{}
	if err != nil {
		return err
	}
	// Setting an empty UUID to trigger the generation of a new one.
	nib.UUID = uuid
	nib.AppendRevision(rev)
	//FIXME: timestamp, deviceID etc.
	nibStore, err := r.getNIBStore()
	if err != nil {
		return err
	}

	return nibStore.Add(&nib)
}

// getFilesNIBUUID returns the NIB for the given relative path or
// returns ErrNoNIB if no pre-existing NIB can be found.
func (r *Repository) getFilesNIBUUID(relPath string) (string, error) {
	//FIXME: implement after deciding on #67
	return "", nil
}

// AddObject adds an object into the storage with the given
// id and adds the data in the reader to it.
func (r *Repository) AddObject(objectID string, data io.Reader) error {
	storage, err := r.getStorage()
	if err != nil {
		return err
	}
	return storage.Set(objectID, data)
}

// AddNIBContent adds NIBData to the repository after verifying it.
func (r *Repository) AddNIBContent(UUID string, nibData io.Reader) error {
	nibStore, err := r.getNIBStore()
	if err != nil {
		return err
	}

	err = nibStore.VerifyContent(nibData)
	if err != nil {
		return err
	}

	return nibStore.AddContent(UUID, nibData)
}

// GetNIB returns a NIB for the given UUID in this repository.
func (r *Repository) GetNIB(UUID string) (*NIB, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}

	return store.Get(UUID)
}

// GetNIBReader returns the NIB with the given UUID in this repository.
func (r *Repository) GetNIBReader(UUID string) (io.Reader, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}

	return store.GetReader(UUID)
}

// GetNIBsFrom returns nibs added since the passed transaction ID.
func (r *Repository) GetNIBsFrom(fromTransactionId int64) (<-chan *NIB, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}
	return store.GetFrom(fromTransactionId)
}

func (r *Repository) GetAllNibs() (<-chan *NIB, error) {
	store, err := r.getNIBStore()
	if err != nil {
		return nil, err
	}
	return store.GetAll()
}

// GetObjectData returns the data stored for the given objectID in this
// repository.
func (r *Repository) GetObjectData(objectID string) (io.Reader, error) {
	storage, err := r.getStorage()
	if err != nil {
		return nil, err
	}
	return storage.Get(objectID)
}

// HasObject returns if the given objectID exists in this repository.
func (r *Repository) HasObject(objectID string) bool {
	storage, err := r.getStorage()
	if err != nil {
		return false
	}

	return storage.Exists(objectID)
}

// hasUUID checks if the given UUID is already in use in this repository;
// this is a local-only check.
func (r *Repository) hasUUID(hash []byte) (bool, error) {
	UUID := formatUUID(hash)
	nibStore, err := r.getNIBStore()
	if err != nil {
		return false, err
	}
	return nibStore.Exists(UUID), nil
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
		Type:             MetadataTypeFile, //FIXME
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
	var enc []byte

	// first generate and encrypt the per-file key and append it to
	// the result buffer:
	var nonce1 [24]byte
	_, err := rand.Read(nonce1[:])
	if err != nil {
		return nil, err
	}

	var fileKey [32]byte
	_, err = rand.Read(fileKey[:])
	if err != nil {
		return nil, err
	}
	repoKey, err := r.GetEncryptionKey()
	if err != nil {
		return nil, err
	}
	out := secretbox.Seal(enc, fileKey[:], &nonce1, &repoKey)

	// then append the actual encrypted contents
	var nonce2 [24]byte
	_, err = rand.Read(nonce2[:])
	if err != nil {
		return nil, err
	}
	out = secretbox.Seal(out, data, &nonce2, &fileKey)
	return out, nil
}

// hashChunk takes a chunk of data and constructs its content-addressing
// hash.
func (r *Repository) hashChunk(chunk []byte) (string, error) {
	key, err := r.GetHashingKey()
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
