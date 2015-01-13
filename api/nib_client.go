package api

import (
	"io"
	"net/http"
)

// putNIBRequest builds a request for uploading NIB
func (c *Client) putNIBRequest(nibID string, nibReader io.ReadCloser) (*http.Request, error) {
	req, err := http.NewRequest("PUT", c.BaseURL+"/nibs/"+nibID,
		nibReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	SignWithKey(req, c.signingPrivateKey)
	return req, nil
}

// PutNIB uploads a NIB to the server
func (c *Client) PutNIB(nibID string, nibReader io.ReadCloser) error {
	req, err := c.putNIBRequest(nibID, nibReader)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, http.StatusCreated, http.StatusOK)
	return err
}
