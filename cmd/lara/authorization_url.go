package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
)

type AuthorizationURL struct {
	URL     *url.URL
	SignKey [PrivateKeySize]byte
	EncKey  [EncryptionKeySize]byte
}

func parseAuthURL(URL *url.URL) (*AuthorizationURL, error) {
	authURL := &AuthorizationURL{}
	err := authURL.parse(URL)
	if err != nil {
		return nil, err
	}
	return authURL, nil
}

func newAuthURL(repositoryBaseUrl string,
	signingPrivKey *[PrivateKeySize]byte,
	encryptionKey *[EncryptionKeySize]byte) (*AuthorizationURL, error) {

	pubKey := edhelpers.GetPublicKeyFromPrivate(*signingPrivKey)
	pubKeyString := hex.EncodeToString(pubKey[:])

	u, err := url.Parse(fmt.Sprintf("%s/authorizations/%s", repositoryBaseUrl, pubKeyString))
	if err != nil {
		return nil, err
	}

	authURL := &AuthorizationURL{
		SignKey: *signingPrivKey,
		EncKey:  *encryptionKey,
		URL:     u,
	}
	return authURL, nil
}

func (a *AuthorizationURL) SignKeyString() string {
	return hex.EncodeToString(a.SignKey[:])
}

func (a *AuthorizationURL) EncKeyString() string {
	return hex.EncodeToString(a.EncKey[:])
}

func (a *AuthorizationURL) String() string {
	return fmt.Sprintf("%s#AuthEncKey=%s&AuthSignKey=%s",
		a.URL.String(), a.EncKeyString(), a.SignKeyString())
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

	a.URL = URL
	a.SignKey = signKey
	a.EncKey = encKey
	return nil
}

func (a *AuthorizationURL) parseForEncKey(data string) ([EncryptionKeySize]byte, error) {
	encKeyRegexp := regexp.MustCompile("AuthEncKey=(?P<key>[^&]+)")
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

func (a *AuthorizationURL) parseForSignKey(data string) ([PrivateKeySize]byte, error) {
	signKeyRegexp := regexp.MustCompile("AuthSignKey=(?P<key>[^&]+)")
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