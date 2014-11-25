package request

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go.crypto/pbkdf2"

	"github.com/hoffie/lara/helpers"
)

const (
	// SaltLen is the length of our signing salts
	SaltLen = 8
)

// SignAsAdmin signs the given request using the admin-shared-secret approach.
func SignAsAdmin(req *http.Request, secret []byte) error {
	salt, err := genSalt()
	if err != nil {
		return err
	}
	key := secretToKey(salt, secret)
	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}
	hash := hmacSign(req, key)
	req.Header.Set("Authorization",
		fmt.Sprintf("lara admin %s%s",
			hex.EncodeToString(salt),
			hex.EncodeToString(hash)))
	return nil
}

// ValidateAdminSigned checks whether the request is signed using the
// admin-shared-secret approach, whether the signature is correct and whether
// the request is not outdated according to the provided maxAge.
func ValidateAdminSigned(req *http.Request, secret []byte, maxAge time.Duration) bool {
	if !validateAdminSig(req, secret) {
		return false
	}
	if !youngerThan(req, maxAge) {
		return false
	}
	return true
}

// adminValidateSig is a helper which ensures that the request's signature
// is an admin signature and is valid.
func validateAdminSig(req *http.Request, secret []byte) bool {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	if !strings.HasPrefix(auth, "lara admin ") {
		return false
	}
	hash := strings.TrimPrefix(auth, "lara admin ")
	if hash == "" {
		return false
	}
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}
	if len(hashBytes) < SaltLen {
		return false
	}
	salt := hashBytes[:SaltLen]
	hashBytes = hashBytes[SaltLen:]
	key := secretToKey(salt, secret)
	realHash := hmacSign(req, key)
	if !helpers.ConstantTimeBytesEqual(hashBytes, realHash) {
		return false
	}
	return true
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

// hmacSign returns the requests HMAC
func hmacSign(req *http.Request, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	concatenateTo(req, mac)
	return mac.Sum(nil)
}

// genSalt generates a new cryptographical salt.
func genSalt() ([]byte, error) {
	salt := make([]byte, SaltLen)
	read, err := rand.Read(salt)
	if read != SaltLen || err != nil {
		return nil, errors.New("unable to generate salt")
	}
	return salt, nil
}

// secretToKey converts the user-supplied password to a key, usable for
// further signing purposes.
func secretToKey(salt, secret []byte) []byte {
	//PERFORMANCE: 4096 as a work factor may be too high as this runs per-request
	key := pbkdf2.Key(secret, salt, 4096, sha256.Size, sha256.New)
	return key
}
