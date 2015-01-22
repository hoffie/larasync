package client

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/hoffie/larasync/api/common"
	"github.com/hoffie/larasync/helpers/crypto"
	"github.com/hoffie/larasync/repository"
)

// putAuthorizationRequest generates a request to add new authorization
// data to the server
func (c *Client) putAuthorizationRequest(
	pubKey *[PublicKeySize]byte,
	authorizationReader io.Reader,
) (*http.Request, error) {
	pubKeyString := hex.EncodeToString(pubKey[:])

	req, err := http.NewRequest(
		"PUT",
		c.BaseURL+"/authorizations/"+pubKeyString,
		authorizationReader,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	common.SignWithKey(req, c.signingPrivateKey)
	return req, nil
}

// PutAuthorization adds a new authorization assignment
// for the passed public key to the server.
func (c *Client) PutAuthorization(
	pubKey *[PublicKeySize]byte,
	authorizationReader io.Reader,
) error {
	req, err := c.putAuthorizationRequest(pubKey, authorizationReader)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, http.StatusOK, http.StatusCreated)
	return err
}

// getAuthorizationRequest generates a request to request a authorization
// from the server.
func (c *Client) getAuthorizationRequest(authorizationURL string,
	authPrivKey [PrivateKeySize]byte) (*http.Request, error) {
	req, err := http.NewRequest(
		"GET",
		authorizationURL,
		nil,
	)
	if err != nil {
		return nil, err
	}

	common.SignWithKey(req, authPrivKey)
	return req, nil
}

// GetAuthorization loads the authorizaion from a server as a URL. Authenticates
// itself against the server with the passed authorization key.
func (c *Client) getAuthorization(authorizationURL string,
	authPrivKey [PrivateKeySize]byte) (io.Reader, error) {
	req, err := c.getAuthorizationRequest(authorizationURL, authPrivKey)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ImportAuthorization generates a new repository "repoName" and imports the
// authorization information from the given URL.
func ImportAuthorization(repoName string, urlString string) (*Client, *repository.ClientRepository, error) {
	repo := repository.NewClient(repoName)
	err := repo.Create()
	if err != nil && !os.IsExist(err) {
		return nil, nil, fmt.Errorf("repository creation failure (%s)", err)
	}

	u, err := url.Parse(urlString)
	if err != nil {
		return nil, nil, fmt.Errorf("unparsable url (%s)", err)
	}
	authURL, err := parseAuthURL(u)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to extract authorization information (%s)", err)
	}

	sc, err := repo.StateConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to load state config (%s)\n", err)
	}

	defaultServer := sc.DefaultServer
	defaultServer.URL = "https://" + u.Host + path.Dir(path.Dir(u.Path))
	defaultServer.Fingerprint = authURL.Fingerprint
	err = sc.Save()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to save state config (%s)\n", err)
	}

	c := New(defaultServer.URL, defaultServer.Fingerprint, nil)

	reader, err := c.getAuthorization(authURL.URL.String(), authURL.SignKey)
	if err != nil {
		return nil, nil, fmt.Errorf("server communication failure (%s)", err)
	}

	enc, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("server data retrieval failed (%s)", err)
	}

	box := crypto.NewBox(authURL.EncKey)
	data, err := box.DecryptContent(enc)
	if err != nil {
		return nil, nil, fmt.Errorf("response decryption failure (%s)", err)
	}

	auth := &repository.Authorization{}
	_, err = auth.ReadFrom(bytes.NewBuffer(data))
	if err != nil {
		return nil, nil, fmt.Errorf("authorization data read failure (%s)", err)
	}

	err = repo.SetKeysFromAuth(auth)
	if err != nil {
		return nil, nil, fmt.Errorf("key storage failure (%s)", err)
	}

	privKey, err := repo.GetSigningPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("private signing key retrieval failure (%s)", err)
	}
	c.SetSigningPrivateKey(privKey)

	return c, repo, nil
}
