package client

import (
	"net/http"

	"github.com/hoffie/larasync/api/tls"
)

// Client provides convenience methods for accessing an api.Server
// over HTTP.
type Client struct {
	http              *http.Client
	BaseURL           string
	netloc            string
	adminSecret       []byte
	signingPrivateKey [PrivateKeySize]byte
}

// NetlocToURL returns the URL matching the given netloc
func NetlocToURL(netloc, repoName string) string {
	// IMPROVEMENT: use mux router to generate URLs
	return "https://" + netloc + "/repositories/" + repoName
}

// New returns a new Client instance.
func New(url, fingerprint string, fingerprintVerifier tls.VerificationFunc) *Client {
	fpv := tls.FingerprintVerifier{
		AcceptFingerprint: fingerprint,
		VerificationFunc:  fingerprintVerifier,
	}
	tr := &http.Transport{
		DialTLS: fpv.DialTLS,
	}
	return &Client{
		http:    &http.Client{Transport: tr},
		BaseURL: url,
	}
}

// SetAdminSecret sets the admin secret to use (e.g. for Register()).
func (c *Client) SetAdminSecret(s []byte) {
	c.adminSecret = s
}

// SetSigningPrivateKey sets the signing private key to use
func (c *Client) SetSigningPrivateKey(k [PrivateKeySize]byte) {
	c.signingPrivateKey = k
}

// doRequest executes the given request and verifies the resulting status code
func (c *Client) doRequest(req *http.Request, expStatus ...int) (*http.Response, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	for _, expStatus := range expStatus {
		if resp.StatusCode == expStatus {
			return resp, nil
		}
	}
	Log.Error("unexpected status", "got", resp.StatusCode, "wanted", expStatus)
	return nil, ErrUnexpectedStatus
}
