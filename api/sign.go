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
	PubkeySize    = ed25519.PublicKeySize
	SignatureSize = ed25519.SignatureSize
)

var staticSalt = []byte("larasync")

// SignAsAdmin signs the given request using the admin-shared-secret approach.
func SignAsAdmin(req *http.Request, secret []byte) {
	key := secretToKey(secret)
	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}
	hash := asymmetricSign(req, key)
	req.Header.Set("Authorization",
		fmt.Sprintf("lara admin %s",
			hex.EncodeToString(hash)))
	return
}

// ValidateAdminSigned checks whether the request is signed using the
// admin-shared-secret approach, whether the signature is correct and whether
// the request is not outdated according to the provided maxAge.
func ValidateAdminSigned(req *http.Request, pubkey [PubkeySize]byte, maxAge time.Duration) bool {
	if !validateAdminSig(req, pubkey) {
		return false
	}
	if !youngerThan(req, maxAge) {
		return false
	}
	return true
}

// adminValidateSig is a helper which ensures that the request's signature
// is an admin signature and is valid.
func validateAdminSig(req *http.Request, pubkey [PubkeySize]byte) bool {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	if !strings.HasPrefix(auth, "lara admin ") {
		return false
	}
	sig := strings.TrimPrefix(auth, "lara admin ")
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
	return asymmetricVerify(req, pubkey, *sigArr)
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

// asymmetricSign uses public key cryptography to sign the request and return the
// signature.
func asymmetricSign(req *http.Request, key [ed25519.PrivateKeySize]byte) []byte {
	mac := sha512.New()
	concatenateTo(req, mac)
	hash := mac.Sum(nil)
	sig := ed25519.Sign(&key, hash)
	slSig := make([]byte, len(sig))
	copy(slSig, sig[0:len(sig)])
	return slSig
}

func asymmetricVerify(req *http.Request, pubkey [PubkeySize]byte, sig [SignatureSize]byte) bool {
	mac := sha512.New()
	concatenateTo(req, mac)
	hash := mac.Sum(nil)
	return ed25519.Verify(&pubkey, hash, &sig)
}

// secretToKey converts the user-supplied password to a key, usable for
// further signing purposes.
func secretToKey(secret []byte) [ed25519.PrivateKeySize]byte {
	//PERFORMANCE/SECURITY: 4096 as a work factor may have to be adapted (runs per request)
	key := pbkdf2.Key(secret, staticSalt, 4096, sha512.Size, sha512.New)
	reader := bytes.NewBuffer(key)
	_, priv, _ := ed25519.GenerateKey(reader)
	return *priv
}

func GetAdminSecretPubkey(secret []byte) [PubkeySize]byte {
	key := secretToKey(secret)
	return edhelpers.GetPublicKeyFromPrivate(key)
}
