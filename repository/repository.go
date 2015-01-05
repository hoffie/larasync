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
)

const (
	authPubkeyFileName     = "auth.pub"
	encryptionKeyFileName  = "encryption.key"
	hashingKeyFileName     = "hashing.key"
	signingPrivkeyFileName = "signing.priv"
	managementDirName      = ".lara"
	objectsDirName         = "objects"
	nibsDirName            = "nibs"
	defaultFilePerms       = 0600
	defaultDirPerms        = 0700
	// EncryptionKeySize represents the size of the key used for
	// encrypting.
	EncryptionKeySize = 32
	// HashingKeySize represents the size of the key used for
	// generating content hashes (HMAC).
	HashingKeySize = 32
)

// Repository represents an on-disk repository and provides methods to
// access its sub-items.
type Repository struct {
	Path    string
	storage BlobStorage
}

// New returns a new repository instance with the given base path
func New(path string) *Repository {
	return &Repository{Path: path}
}

// getStorage returns the currently configured blob storage backend
// for the repository.
func (r *Repository) getStorage() (BlobStorage, error) {
	if r.storage == nil {
		storage := FileBlobStorage{
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

// CreateManagementDir ensures that this repository's management
// directory exists.
func (r *Repository) CreateManagementDir() error {
	err := os.Mkdir(r.GetManagementDir(), defaultDirPerms)
	if err != nil && !os.IsExist(err) {
		return err
	}
	path := r.getNIBsPath()
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

// getAuthPubkeyPath returns the path of the repository's auth pubkey
// storage location.
func (r *Repository) getAuthPubkeyPath() string {
	return filepath.Join(r.Path, managementDirName, authPubkeyFileName)
}

// getEncryptionKeyPath returns the path of the repository's encryption key
// storage location.
func (r *Repository) getEncryptionKeyPath() string {
	return filepath.Join(r.Path, managementDirName, encryptionKeyFileName)
}

// getSigningPrivkeyPath returns the path of the repository's signing
// private key location.
func (r *Repository) getSigningPrivkeyPath() string {
	return filepath.Join(r.Path, managementDirName, signingPrivkeyFileName)
}

// getHashingKeyPath returns the path of the repository's hashing
// key location.
func (r *Repository) getHashingKeyPath() string {
	return filepath.Join(r.Path, managementDirName, hashingKeyFileName)
}

// getObjectsPath returns the path of the repository's objects location
func (r *Repository) getObjectsPath() string {
	return filepath.Join(r.Path, managementDirName, objectsDirName)
}

// getNibsDirName returns the path of the repository's nibs location
func (r *Repository) getNIBsPath() string {
	return filepath.Join(r.Path, managementDirName, nibsDirName)
}

// GetAuthPubkey returns the repository auth key's public key.
func (r *Repository) GetAuthPubkey() ([]byte, error) {
	pubkey, err := ioutil.ReadFile(r.getAuthPubkeyPath())
	return pubkey, err
}

// SetAuthPubkey sets the repository auth key's public key.
func (r *Repository) SetAuthPubkey(key []byte) error {
	return ioutil.WriteFile(r.getAuthPubkeyPath(), key, defaultFilePerms)
}

// SetEncryptionKey sets the repository encryption key
func (r *Repository) SetEncryptionKey(key [EncryptionKeySize]byte) error {
	return ioutil.WriteFile(r.getEncryptionKeyPath(), key[:], defaultFilePerms)
}

// GetEncryptionKey returns the repository encryption key.
func (r *Repository) GetEncryptionKey() ([EncryptionKeySize]byte, error) {
	key, err := ioutil.ReadFile(r.getEncryptionKeyPath())
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
	return ioutil.WriteFile(r.getSigningPrivkeyPath(), key[:], defaultFilePerms)
}

// GetSigningPrivkey returns the repository signing private key.
func (r *Repository) GetSigningPrivkey() ([PrivateKeySize]byte, error) {
	key, err := ioutil.ReadFile(r.getSigningPrivkeyPath())
	if len(key) != PrivateKeySize {
		return [PrivateKeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [PrivateKeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
}

// SetHashingKey sets the repository hashing key (content addressing)
func (r *Repository) SetHashingKey(key [HashingKeySize]byte) error {
	return ioutil.WriteFile(r.getHashingKeyPath(), key[:], defaultFilePerms)
}

// GetHashingKey returns the repository signing private key.
func (r *Repository) GetHashingKey() ([HashingKeySize]byte, error) {
	key, err := ioutil.ReadFile(r.getHashingKeyPath())
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

	//FIXME this only works for new files / non-existing NIBs
	rev := &Revision{}
	rev.MetadataID = metadataID
	rev.ContentIDs = contentIDs
	nib := NIB{}
	uuid, err := r.findFreeUUID()
	if err != nil {
		return err
	}
	nib.UUID = string(uuid)
	nib.AppendRevision(rev)
	//FIXME: timestamp, deviceID etc.
	buf := &bytes.Buffer{}
	_, err = nib.WriteTo(buf)
	if err != nil {
		return err
	}
	err = r.writeNIB(formatUUID(uuid), buf.Bytes())
	return err
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

// findFreeUUID generates a new UUID for naming a NIB; it tries to avoid
// local collisions.
func (r *Repository) findFreeUUID() ([]byte, error) {
	hostname := os.Getenv("HOSTNAME")
	rnd := make([]byte, 32)
	for {
		_, err := rand.Read(rnd)
		if err != nil {
			return nil, err
		}
		hasher := sha512.New()
		hasher.Write([]byte(hostname))
		hasher.Write(rnd)
		hash := hasher.Sum(nil)
		hasUUID, err := r.hasUUID(hash)
		if err != nil {
			return nil, err
		}
		if !hasUUID {
			return hash, nil
		}
	}
	return nil, errors.New("findFreeUUID: unexpected case")
}

// hasUUID checks if the given UUID is already in use in this repository;
// this is a local-only check.
func (r *Repository) hasUUID(hash []byte) (bool, error) {
	hexHash := hex.EncodeToString(hash)
	path := filepath.Join(r.getNIBsPath(), hexHash)
	s, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if s.IsDir() {
		return false, errors.New("is directory")
	}
	return true, nil
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

// writeNIB writes a node information block to disk.
// It signs the data in the process.
func (r *Repository) writeNIB(name string, data []byte) error {
	path := filepath.Join(r.getNIBsPath(), name)
	key, err := r.GetSigningPrivkey()
	if err != nil {
		return err
	}
	w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		defaultFilePerms)
	if err != nil {
		return err
	}
	defer w.Close()

	sw := NewSigningWriter(key, w)
	_, err = sw.Write(data)
	if err != nil {
		return err
	}
	err = sw.Finalize()
	return err
}

// hashChunk takes a chunk of data and constructs its content-addressing
// hash.
func (r *Repository) hashChunk(chunk []byte) (string, error) {
	key, err := r.GetHashingKey()
	if err != nil {
		return "", err
	}
	hasher := hmac.New(sha512.New, key[:])
	hash := hasher.Sum(chunk)
	hexHash := hex.EncodeToString(hash)
	return hexHash, nil
}

// writeFileToChunks takes a file path and saves its contents to the
// storage in encrypted form with a content-addressing id.
func (r *Repository) writeFileToChunks(path string) ([]string, error) {
	//FIXME split in chunks
	chunk, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	id, err := r.writeContentAddressedCryptoContainer(chunk)
	if err != nil {
		return nil, err
	}
	return []string{id}, nil
}
