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

// getObjectRequest builds a request for downloading an object
func (c *Client) getObjectRequest(objectID string) (*http.Request, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/blobs/"+objectID, nil)
	if err != nil {
		return nil, err
	}
	SignWithKey(req, c.signingPrivateKey)
	return req, nil
}

// GetObject downloads an object from the server
func (c *Client) GetObject(objectID string) (io.Reader, error) {
	req, err := c.getObjectRequest(objectID)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req, http.StatusCreated, http.StatusOK)
	if resp == nil {
		return nil, err
	}
	return resp.Body, err
}
