package api

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
)

// concatenatorDelimiter is the character which is put between different
// parts of the request so that these parts cannot be mixed up.
const concatenatorDelimiter = byte(0)

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
	c.writeDelimitedString(c.req.Method)
	c.URL()
	c.Headers()
	// Body() must be last, as it does not use the length-limited bincontainer;
	// instead, it writes its data verbatim
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

	c.writeDelimitedString(url.String())
}

// ignoreHeaders is a list of all headers which are not part of the
// resulting output.
var ignoreHeaders = map[string]bool{
	// Authorization is our header; we can't sign our own signature
	"Authorization": true,
	// User-Agent is allowed to vary
	"User-Agent": true,
	// Accept-Encoding will also be mangled by the client
	"Accept-Encoding": true,
	// Content-Length doesn't matter as the content is signed
	"Content-Length": true,
	// Host header does not need to be included as the parsed host header
	// is included as part of the URL (see URL())
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
		c.writeDelimitedString(header)
		for _, value := range c.req.Header[header] {
			c.writeDelimitedString(value)
		}
	}
}

// Body concatenates the body.
func (c *concatenator) Body() error {
	if c.req.Body == nil {
		return c.writeDelimiter()
	}
	bodyCopy := &bytes.Buffer{}
	// we write directly to the writer here and do not use
	// bincontainer.Encoder; the reason for this is that doing
	// that would require buffering all the body content to
	// calculate its length beforehand.
	// for this to be safe, no other data may be written afterwards.
	_, err := io.Copy(io.MultiWriter(bodyCopy, c.w), c.req.Body)
	if err != nil {
		return err
	}
	c.req.Body = ioutil.NopCloser(bodyCopy)
	return c.writeDelimiter()
}

// writeDelimitedString writes the given string to the writer and
// appends the delimiter.
func (c *concatenator) writeDelimitedString(s string) error {
	_, err := c.w.Write([]byte(s))
	if err != nil {
		return err
	}
	return c.writeDelimiter()
}

// writeDelimiter outputs the delimiter to the writer.
func (c *concatenator) writeDelimiter() error {
	_, err := c.w.Write([]byte{concatenatorDelimiter})
	return err
}
