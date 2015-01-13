package api

import (
	"io"
	"net/http"
)

// putObjectRequest builds a request for uploading an object
func (c *Client) putObjectRequest(objectID string, content io.ReadCloser) (*http.Request, error) {
	req, err := http.NewRequest("PUT", c.BaseURL+"/blobs/"+objectID,
		content)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	SignWithKey(req, c.signingPrivateKey)
	return req, nil
}

// PutObject uploads an object to the server
func (c *Client) PutObject(objectID string, content io.ReadCloser) error {
	req, err := c.putObjectRequest(objectID, content)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, http.StatusCreated, http.StatusOK)
	return err
}
