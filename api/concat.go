package api

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
)

type concatenator struct {
	req *http.Request
	w   io.Writer
}

// concatenateTo writes all elementary request parts to the given writer.
// parts are: method, URL, headers, body
func concatenateTo(req *http.Request, w io.Writer) {
	c := &concatenator{
		req: req,
		w:   w,
	}
	c.Basics()
	c.Headers()
	c.Body()
}

// Basics concatenates method and URL.
func (c *concatenator) Basics() {
	c.w.Write([]byte(c.req.Method))
	c.w.Write([]byte(c.req.URL.String()))
}

// Headers concatenate the headers.
func (c *concatenator) Headers() {
	headers := make([]string, len(c.req.Header))
	i := 0
	for header := range c.req.Header {
		headers[i] = header
		i++
	}
	sort.Strings(headers)
	for _, header := range headers {
		if header == "Authorization" {
			continue
		}
		c.w.Write([]byte(header))
		for _, value := range c.req.Header[header] {
			c.w.Write([]byte(value))
		}
	}
}

// Body concatenates the body.
func (c *concatenator) Body() {
	if c.req.Body == nil {
		return
	}
	bodyCopy := &bytes.Buffer{}
	io.Copy(io.MultiWriter(bodyCopy, c.w), c.req.Body)
	c.req.Body = ioutil.NopCloser(bodyCopy)
}
