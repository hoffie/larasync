package repository

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hoffie/larasync/helpers/atomic"
	"github.com/hoffie/larasync/helpers/crypto"
)

const (
	// internal directory names
	managementDirName     = ".lara"
	objectsDirName        = "objects"
	nibsDirName           = "nibs"
	transactionsDirName   = "transactions"
	authorizationsDirName = "authorizations"
	stateConfigFileName   = "state.json"

	// default permissions
	defaultFilePerms = 0600
	defaultDirPerms  = 0700

	// chunk splitting size
	chunkSize = 1 * 1024 * 1024
)

// Repository represents an on-disk repository and provides methods to
// access its sub-items.
type Repository struct {
	Path                 string
	keys                 *KeyStore
	objectStorage        ContentStorage
	nibStore             *NIBStore
	transactionManager   *TransactionManager
	authorizationManager *AuthorizationManager
	stateConfig          *StateConfig
	managementDirPath    string
}

// New returns a new repository instance with the given base path
func New(path string) *Repository {
	r := &Repository{Path: path}
	r.setupPaths()

	r.objectStorage = newFileContentStorage(r.subPathFor(objectsDirName))

	r.transactionManager = newTransactionManager(
		newFileContentStorage(r.subPathFor(transactionsDirName)),
		r.GetManagementDir(),
	)
	r.authorizationManager = newAuthorizationManager(
		newFileContentStorage(r.subPathFor(authorizationsDirName)),
	)

	r.keys = NewKeyStore(r.managementDirPath)
	r.nibStore = newNIBStore(
		newFileContentStorage(r.subPathFor(nibsDirName)),
		r.keys,
		r.transactionManager,
	)

	return r
}

// subPathFor returns the full path for the given entry.
func (r *Repository) subPathFor(name string) string {
	return filepath.Join(r.GetManagementDir(), name)
}

// setupPaths initializes several attributes referring to internal repository paths
// such as encryption key paths.
func (r *Repository) setupPaths() {
	base := filepath.Join(r.Path, managementDirName)
	r.managementDirPath = base
}

// CreateManagementDir ensures that this repository's management
// directory exists.
func (r *Repository) CreateManagementDir() error {
	err := os.Mkdir(r.GetManagementDir(), defaultDirPerms)
	if err != nil && !os.IsExist(err) {
		return err
	}

	storages := []*FileContentStorage{
		newFileContentStorage(r.subPathFor(authorizationsDirName)),
		newFileContentStorage(r.subPathFor(nibsDirName)),
		newFileContentStorage(r.subPathFor(transactionsDirName)),
		newFileContentStorage(r.subPathFor(objectsDirName)),
	}

	for _, fileStorage := range storages {
		err = fileStorage.CreateDir()
		if err != nil {
			return err
		}
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
	if err != nil {
		return err
	}

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
	if err != nil {
		writer.Close()
		return err
	}

	for _, contentID := range rev.ContentIDs {
		content, err := r.readEncryptedObject(contentID)
		_, err = writer.Write(content)
		if err != nil {
			return err
		}
	}

	hasChanges, err := r.pathHasConflictingChanges(nib, absPath)
	if err != nil {
		return err
	}
	if hasChanges {
		return errors.New("workdir conflict")
	}

	err = writer.Close()
	if err != nil {
		return err
	}
	return nil
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

// getFilesNIBUUID returns the NIB for the given relative path
func (r *Repository) pathToNIBID(relPath string) (string, error) {
	return r.hashChunk([]byte(relPath))
}

// AddObject adds an object into the storage with the given
// id and adds the data in the reader to it.
func (r *Repository) AddObject(objectID string, data io.Reader) error {
	return r.objectStorage.Set(objectID, data)
}

// HasObject returns whether the given object id exists in the object
// store.
func (r *Repository) HasObject(objectID string) bool {
	return r.objectStorage.Exists(objectID)
}

// AddNIBContent adds NIBData to the repository after verifying it.
func (r *Repository) AddNIBContent(nibReader io.Reader) error {
	nibStore := r.nibStore

	data, err := ioutil.ReadAll(nibReader)
	if err != nil {
		return err
	}

	nib, err := nibStore.VerifyAndParseBytes(data)
	if err != nil {
		return err
	}

	return nibStore.AddContent(nib.ID, bytes.NewReader(data))
}

// GetNIB returns a NIB for the given ID in this repository.
func (r *Repository) GetNIB(id string) (*NIB, error) {
	return r.nibStore.Get(id)
}

// GetNIBReader returns the NIB with the given id in this repository.
func (r *Repository) GetNIBReader(id string) (io.ReadCloser, error) {
	return r.nibStore.getReader(id)
}

// GetNIBBytesFrom returns the signed byte structure for NIBs from the given
// transaction id
func (r *Repository) GetNIBBytesFrom(fromTransactionID int64) (<-chan []byte, error) {
	return r.nibStore.GetBytesFrom(fromTransactionID)
}

// GetNIBsFrom returns nibs added since the passed transaction ID.
func (r *Repository) GetNIBsFrom(fromTransactionID int64) (<-chan *NIB, error) {
	return r.nibStore.GetFrom(fromTransactionID)
}

// GetAllNIBBytes returns all NIBs signed byte representations in this repository.
func (r *Repository) GetAllNIBBytes() (<-chan []byte, error) {
	return r.nibStore.GetAllBytes()
}

// GetAllNibs returns all the nibs which are stored in this repository.
// Those will be returned with the oldest one first and the newest added
// last.
func (r *Repository) GetAllNibs() (<-chan *NIB, error) {
	return r.nibStore.GetAll()
}

// HasNIB checks if a NIB with the given ID exists in the repository.
func (r *Repository) HasNIB(id string) bool {
	return r.nibStore.Exists(id)
}

// CurrentTransaction returns the currently newest Transaction for this
// repository.
func (r *Repository) CurrentTransaction() (*Transaction, error) {
	return r.transactionManager.CurrentTransaction()
}

// GetAuthorizationReader returns the authorization configuration for the
// passed PublicKey.
func (r *Repository) GetAuthorizationReader(publicKey [PublicKeySize]byte) (io.ReadCloser, error) {
	return r.authorizationManager.GetReader(publicKey)
}

// SetAuthorizationData adds for the given publicKey the authorization structure
func (r *Repository) SetAuthorizationData(publicKey [PublicKeySize]byte, authData io.Reader) error {
	return r.authorizationManager.SetData(publicKey, authData)
}

// SetAuthorization adds a authorization with the given publicKey and encrypts it with the
// passed encryptionKey to this repository.
func (r *Repository) SetAuthorization(
	publicKey [PublicKeySize]byte,
	encKey [EncryptionKeySize]byte,
	authorization *Authorization,
) error {
	return r.authorizationManager.Set(publicKey, encKey, authorization)
}

// DeleteAuthorization removes the authorization with the given publicKey.
func (r *Repository) DeleteAuthorization(publicKey [PublicKeySize]byte) error {
	return r.authorizationManager.Delete(publicKey)
}

// SerializeAuthorization returns the encrypted and authorization which can be passed
// safely to the server.
func (r *Repository) SerializeAuthorization(encryptionKey [EncryptionKeySize]byte,
	authorization *Authorization) ([]byte, error) {
	return r.authorizationManager.Serialize(encryptionKey, authorization)
}

// CurrentAuthorization returns the currently valid Authorization object
// for this repository. If the privateKeys necessary for this are not
// stored in the keyStore (common case if this is a server) an error is
// returned.
func (r *Repository) CurrentAuthorization() (*Authorization, error) {
	keys := r.keys
	encryptionKey, err := keys.EncryptionKey()
	if err != nil {
		return nil, errors.New("Could not load encryption key.")
	}

	hashingKey, err := keys.HashingKey()
	if err != nil {
		return nil, errors.New("Could not load hashing key.")
	}

	signatureKey, err := keys.SigningPrivateKey()
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

// GetObjectData returns the data stored for the given objectID in this
// repository.
func (r *Repository) GetObjectData(objectID string) (io.ReadCloser, error) {
	return r.objectStorage.Get(objectID)
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

// GetSigningPublicKey exposes the signing public key as it is required
// in foreign packages such as api.
func (r *Repository) GetSigningPublicKey() ([PublicKeySize]byte, error) {
	return r.keys.SigningPublicKey()
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

// SetKeysFromAuth takes the keys passed through the authorization and puts
// them into the keystore.
func (r *Repository) SetKeysFromAuth(auth *Authorization) error {
	keys := r.keys
	err := keys.SetEncryptionKey(auth.EncryptionKey)
	if err != nil {
		return err
	}
	err = keys.SetHashingKey(auth.HashingKey)
	if err != nil {
		return err
	}
	err = keys.SetSigningPrivateKey(auth.SigningKey)
	if err != nil {
		return err
	}
	return nil
}

// StateConfig returns this repository's state config; it is currently used
// in client repositories only and stores things like the default server.
func (r *Repository) StateConfig() (*StateConfig, error) {
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
