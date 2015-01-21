package client

import (
	"io"
	"net/http"

	"github.com/hoffie/larasync/helpers/bincontainer"
	. "github.com/hoffie/larasync/api/common"
)

// putNIBRequest builds a request for uploading NIB
func (c *Client) putNIBRequest(nibID string, nibReader io.Reader) (*http.Request, error) {
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
func (c *Client) PutNIB(nibID string, nibReader io.Reader) error {
	req, err := c.putNIBRequest(nibID, nibReader)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, http.StatusCreated, http.StatusOK)
	return err
}

// getNIBsRequest builds a request for getting a NIB list
func (c *Client) getNIBsRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/nibs", nil)
	if err != nil {
		return nil, err
	}
	SignWithKey(req, c.signingPrivateKey)
	return req, nil
}

// GetNIBs returns the list of all nibs
func (c *Client) GetNIBs() (<-chan []byte, error) {
	req, err := c.getNIBsRequest()
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return nil, err
	}
	bin := bincontainer.NewDecoder(resp.Body)
	res := make(chan []byte, 100)
	go func() {
		for {
			chunk, err := bin.ReadChunk()
			if err == io.EOF {
				close(res)
				return
			}
			if err != nil {
				break
			}
			res <- chunk
		}
		close(res)
	}()
	return res, nil
}
