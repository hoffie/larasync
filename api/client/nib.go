package client

import (
	"io"
	"net/http"
	"strconv"

	"github.com/hoffie/larasync/api/common"
	"github.com/hoffie/larasync/helpers/bincontainer"
)

// putNIBRequest builds a request for uploading NIB
func (c *Client) putNIBRequest(nibID string, nibReader io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("PUT", c.BaseURL+"/nibs/"+nibID,
		nibReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	common.SignWithKey(req, c.signingPrivateKey)
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

// NIBGetResponse encapsulates the response given from a NIB GET request.
type NIBGetResponse struct {
	NIBData             <-chan []byte
	ServerTransactionID int64
}

// getNIBsRequest builds a request for getting a NIB list
func (c *Client) getNIBsRequest() (*http.Request, error) {
	return c.getNIBsFromTransactionRequest(0)
}

// getNIBsFromTransactionRequest builds a request and requests all data
// from the passed Transaction ID.
func (c *Client) getNIBsFromTransactionRequest(lastTransactionID int64) (*http.Request, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/nibs", nil)
	if err != nil {
		return nil, err
	}

	if lastTransactionID != 0 {
		query := req.URL.Query()
		query.Add("from-transaction-id", strconv.FormatInt(lastTransactionID, 10))
		req.URL.RawQuery = query.Encode()
	}

	common.SignWithKey(req, c.signingPrivateKey)
	return req, nil
}

// GetNIBsFromTransactionID returns the list of all nibs from the passed server transaction.
func (c *Client) GetNIBsFromTransactionID(lastTransactionID int64) (*NIBGetResponse, error) {
	req, err := c.getNIBsFromTransactionRequest(lastTransactionID)
	if err != nil {
		return nil, err
	}
	return c.processNibGetRequest(req)
}

// GetNIBs returns the list of all nib byte representations.
func (c *Client) GetNIBs() (*NIBGetResponse, error) {
	req, err := c.getNIBsRequest()
	if err != nil {
		return nil, err
	}
	return c.processNibGetRequest(req)
}

// processNibGetRequest takes a request and processes the NIB GET request. Returns the
// parsed bytes from the entry.
func (c *Client) processNibGetRequest(req *http.Request) (*NIBGetResponse, error) {
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
	nibResponse := &NIBGetResponse{
		NIBData:             res,
		ServerTransactionID: parseTransactionID(resp),
	}
	return nibResponse, nil
}

// parseTransactionID tries to get the current server Transaction ID from the
// passed response. Returns "0" if no transaction ID could be extracted.
func parseTransactionID(resp *http.Response) int64 {
	currentTransactionIDString := resp.Header.Get("X-Current-Transaction-Id")

	currentTransactionID, err := strconv.ParseInt(currentTransactionIDString, 10, 64)
	if err != nil {
		return 0
	}
	return currentTransactionID
}
