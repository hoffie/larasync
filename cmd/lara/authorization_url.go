package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
)

var (
	signKeyRegexp     = regexp.MustCompile("AuthSignKey=(?P<key>[^&]+)")
	encKeyRegexp      = regexp.MustCompile("AuthEncKey=(?P<key>[^&]+)")
	fingerprintRegexp = regexp.MustCompile("Fingerprint=(?P<key>[^&]+)")
)

// AuthorizationURL is used to pass and create a authorizations
// for registering against a new server.
type AuthorizationURL struct {
	URL         *url.URL
	SignKey     [PrivateKeySize]byte
	EncKey      [EncryptionKeySize]byte
	Fingerprint string
}

// parseAuthURL takes a URL and tries to extract the encryption key
// and the auhtorization key from the fragment.
func parseAuthURL(URL *url.URL) (*AuthorizationURL, error) {
	authURL := &AuthorizationURL{}
	err := authURL.parse(URL)
	if err != nil {
		return nil, err
	}
	return authURL, nil
}

// newAuthURL generates a new authorization URL with the passed
// arguments.
func newAuthURL(repositoryBaseURL string,
	signingPrivKey *[PrivateKeySize]byte,
	encryptionKey *[EncryptionKeySize]byte,
	fingerprint string) (*AuthorizationURL, error) {

	pubKey := edhelpers.GetPublicKeyFromPrivate(*signingPrivKey)
	pubKeyString := hex.EncodeToString(pubKey[:])

	u, err := url.Parse(fmt.Sprintf("%s/authorizations/%s", repositoryBaseURL, pubKeyString))
	if err != nil {
		return nil, err
	}

	authURL := &AuthorizationURL{
		SignKey:     *signingPrivKey,
		EncKey:      *encryptionKey,
		Fingerprint: fingerprint,
		URL:         u,
	}
	return authURL, nil
}

// SignKeyString returns the authorization key encoded as hex.
func (a *AuthorizationURL) SignKeyString() string {
	return hex.EncodeToString(a.SignKey[:])
}

// EncKeyString returns the encryption key encoded as hex.
func (a *AuthorizationURL) EncKeyString() string {
	return hex.EncodeToString(a.EncKey[:])
}

// String formats the AuthorizationURL which should be passed to
// the new client to authorize.
func (a *AuthorizationURL) String() string {
	return fmt.Sprintf("%s#AuthEncKey=%s&AuthSignKey=%s&Fingerprint=%s",
		a.URL.String(), a.EncKeyString(), a.SignKeyString(), a.Fingerprint)
}

func (a *AuthorizationURL) parse(URL *url.URL) error {
	authData := URL.Fragment
	URL.Fragment = ""

	encKey, err := a.parseForEncKey(authData)
	if err != nil {
		return err
	}

	signKey, err := a.parseForSignKey(authData)
	if err != nil {
		return err
	}

	fingerprint, err := a.parseForFingerprint(authData)
	if err != nil {
		return err
	}

	a.URL = URL
	a.SignKey = signKey
	a.EncKey = encKey
	a.Fingerprint = fingerprint
	return nil
}

// parseForEncKey tries to extract the encryption key.
func (a *AuthorizationURL) parseForEncKey(data string) ([EncryptionKeySize]byte, error) {
	encKeySlice, err := a.parseForKey(data, encKeyRegexp)
	encKey := [EncryptionKeySize]byte{}
	if err != nil {
		return encKey, errors.New("Could not retrieve encryption key.")
	}
	if len(encKeySlice) != EncryptionKeySize {
		return encKey, errors.New("Invalid encryption key size.")
	}

	copy(encKey[:], encKeySlice)
	return encKey, nil
}

// parseForSignKey tries to extract the signing key.
func (a *AuthorizationURL) parseForSignKey(data string) ([PrivateKeySize]byte, error) {
	signKeySlice, err := a.parseForKey(data, signKeyRegexp)
	signKey := [PrivateKeySize]byte{}
	if err != nil {
		return signKey, errors.New("Could not retrieve signing key.")
	}
	if len(signKeySlice) != PrivateKeySize {
		return signKey, errors.New("Invalid signature key size.")
	}

	copy(signKey[:], signKeySlice)
	return signKey, nil
}

// parseForFingerprint tries to extract the fingerprint
func (a *AuthorizationURL) parseForFingerprint(data string) (string, error) {
	matches := fingerprintRegexp.FindStringSubmatch(data)
	if len(matches) < 2 {
		return "", errors.New("Could not parse fingerprint")
	}
	return string(matches[1]), nil
}

// parseForKey tries to parse a key from the data string and parses it with the given regexp.
func (a *AuthorizationURL) parseForKey(data string, r *regexp.Regexp) ([]byte, error) {
	keyMatches := r.FindStringSubmatch(data)
	extractionError := errors.New("Could not extract key.")

	if len(keyMatches) < 2 {
		return nil, extractionError
	}
	keyString := keyMatches[1]
	keySlice, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, extractionError
	}
	return keySlice, nil
}
