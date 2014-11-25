package request

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"testing"
)

func TestRequestAdminSigning(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.org/repositories", nil)
	if err != nil {
		t.Fatal(err)
	}
	key := []byte("test")

	SignAsAdmin(req, key)
	if req.Header.Get("Authorization") == "" {
		t.Fatal("no authorization header")
	}
	if !ValidateAdminSigned(req, key, time.Minute) {
		buf := bytes.Buffer{}
		concatenateTo(req, &buf)
		t.Log(buf.String())
		t.Fatal("validation failed")
	}

	req.Header.Set("Authorization", "")
	if ValidateAdminSigned(req, key, time.Minute) {
		t.Fatal("validation succeeded even without Authorization header")
	}

	req.Header.Set("Authorization", "basic foo")
	if ValidateAdminSigned(req, key, time.Minute) {
		t.Fatal("validation succeeded even non-lara Authorization header")
	}

	req.Header.Set("Authorization", "lara foo")
	if ValidateAdminSigned(req, key, time.Minute) {
		t.Fatal("validation succeeded even non-admin Authorization header")
	}

	req.Header.Set("Authorization", "lara admin ")
	if ValidateAdminSigned(req, key, time.Minute) {
		t.Fatal("validation succeeded even without hash in Authorization header")
	}

	SignAsAdmin(req, key)
	req.Method = "POST"
	if ValidateAdminSigned(req, key, time.Minute) {
		t.Fatal("validation succeeded even after method change")
	}

	SignAsAdmin(req, key)
	req.Header.Set("Test", "1")
	if ValidateAdminSigned(req, key, time.Minute) {
		t.Fatal("validation succeeded even after header addition")
	}

	SignAsAdmin(req, key)
	newURL, err := url.Parse("http://example.org/repositories?x=3")
	if err != nil {
		t.Fatal(err)
	}
	req.URL = newURL
	if ValidateAdminSigned(req, key, time.Minute) {
		t.Fatal("validation succeeded even after URL changing addition")
	}

	SignAsAdmin(req, key)
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("changed body")))
	if ValidateAdminSigned(req, key, time.Minute) {
		t.Fatal("validation succeeded even after body change")
	}
	// ensure that req.Body is still readable as it would be without signing
	buf := make([]byte, 100)
	read, _ := req.Body.Read(buf)
	if string(buf[:read]) != "changed body" {
		t.Fatal("body no longer readable after signing")
	}

	tenSecsAgo := time.Now().Add(-10 * time.Second)
	// this will most likely send non-GMT time which is against the HTTP RFC;
	// as we should handle this as well, it's ok for testing:
	req.Header.Set("Date", tenSecsAgo.Format(time.RFC1123))
	SignAsAdmin(req, key)
	if ValidateAdminSigned(req, key, 9*time.Second) {
		t.Fatal("validation succeeded even after reaching max signature age")
	}
}
