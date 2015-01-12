package api

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	c.w.Write([]byte(c.req.Method))
	c.URL()
	c.Headers()
	c.Body()
}

// URL (re)constructs the URL and appends it
func (c *concatenator) URL() {
	var url url.URL
	url = *c.req.URL
	// we do not know the scheme for all cases, so ignore it:
	url.Scheme = ""
	if url.Host == "" {
		url.Host = c.req.Host
	}

	c.w.Write([]byte(url.String()))
}

var ignoreHeaders = map[string]bool{
	"Authorization":   true,
	"User-Agent":      true,
	"Accept-Encoding": true,
	// Content-Length doesn't matter as the content is signed
	"Content-Length": true,
	// Host header does not need to be signed as the parsed host header
	// is part of the signature (see URL())
	"Host": true,
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
		_, isIgnored := ignoreHeaders[header]
		if isIgnored {
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
