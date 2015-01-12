package api

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go.crypto/pbkdf2"
	"github.com/agl/ed25519"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
)

const (
	// PrivateKeySize denotes how many bytes a private key needs (binary encoded)
	PrivateKeySize = ed25519.PrivateKeySize
	// PublicKeySize denotes how many bytes a pubkey needs (binary encoded)
	PublicKeySize = ed25519.PublicKeySize
	// SignatureSize denotes how many bytes a sig needs (binary encoded)
	SignatureSize = ed25519.SignatureSize
)

var staticSalt = []byte("larasync")

// SignWithPassphrase signs the given request using the given admin passphrase
func SignWithPassphrase(req *http.Request, passphrase []byte) error {
	key, err := passphraseToKey(passphrase)
	if err != nil {
		return err
	}
	SignWithKey(req, key)
	return nil
}

// SignWithKey signs the request with the given private key by adding an
// appropriate authorization header.
// A Date header is also appended if not yet existing
func SignWithKey(req *http.Request, key [PrivateKeySize]byte) {
	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}
	sig := getSignature(req, key)
	req.Header.Set("Authorization", fmt.Sprintf("lara %s",
		hex.EncodeToString(sig)))
}

// ValidateRequest checks whether the request signature is valid and
// matches the given public key. It also checks whether the request
// is not outdated according to the provided maxAge.
func ValidateRequest(req *http.Request, pubkey [PublicKeySize]byte, maxAge time.Duration) bool {
	if !validateRequestSig(req, pubkey) {
		return false
	}
	if !youngerThan(req, maxAge) {
		return false
	}
	return true
}

// validateRequestSig is a helper which ensures that the request's signature
// is valid. It extracts the signature on its own.
func validateRequestSig(req *http.Request, pubkey [PublicKeySize]byte) bool {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	if !strings.HasPrefix(auth, "lara ") {
		return false
	}
	sig := strings.TrimPrefix(auth, "lara ")
	if sig == "" {
		return false
	}
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	if len(sigBytes) < SignatureSize {
		return false
	}
	sigArr := new([SignatureSize]byte)
	copy(sigArr[:], sigBytes[:SignatureSize])
	return verifySig(req, pubkey, *sigArr)
}

// youngerThan checks whether the request's Date header is at maximum
// maxAge old.
func youngerThan(req *http.Request, maxAge time.Duration) bool {
	dateHeader := req.Header.Get("Date")
	date, err := time.Parse(time.RFC1123, dateHeader)
	if err != nil {
		return false
	}
	if time.Now().UTC().Sub(date) > maxAge {
		return false
	}
	return true
}

// getSignature uses public key cryptography to sign the request
// and return the resulting signature.
func getSignature(req *http.Request, key [PrivateKeySize]byte) []byte {
	mac := sha512.New()
	concatenateTo(req, mac)
	buf := &bytes.Buffer{}
	concatenateTo(req, buf)
	fmt.Println(buf.String())
	hash := mac.Sum(nil)
	sig := ed25519.Sign(&key, hash)
	slSig := make([]byte, len(sig))
	copy(slSig, sig[0:len(sig)])
	return slSig
}

// verifySig checks if the signature matches the provided
// public key and is valid for the given request.
func verifySig(req *http.Request, pubkey [PublicKeySize]byte, sig [SignatureSize]byte) bool {
	mac := sha512.New()
	concatenateTo(req, mac)
	buf := &bytes.Buffer{}
	concatenateTo(req, buf)
	fmt.Println(buf.String())
	hash := mac.Sum(nil)
	return ed25519.Verify(&pubkey, hash, &sig)
}

// passphraseToKey converts the user-supplied passphrase to a key, usable for
// further signing purposes.
func passphraseToKey(passphrase []byte) ([PrivateKeySize]byte, error) {
	//PERFORMANCE/SECURITY: 4096 as a work factor may have to be adapted (runs per request)
	key := pbkdf2.Key(passphrase, staticSalt, 4096, sha512.Size, sha512.New)
	reader := bytes.NewBuffer(key)
	_, priv, err := ed25519.GenerateKey(reader)
	if err != nil {
		return [PrivateKeySize]byte{}, err
	}
	return *priv, nil
}

// GetAdminSecretPubkey transforms the given passphrase into a private key
// and returns the accompying public key (e.g. for storage on the server)
func GetAdminSecretPubkey(passphrase []byte) ([PublicKeySize]byte, error) {
	key, err := passphraseToKey(passphrase)
	if err != nil {
		return [PublicKeySize]byte{}, err
	}
	return edhelpers.GetPublicKeyFromPrivate(key), nil
}
