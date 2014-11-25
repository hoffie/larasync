package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"testing"
)

func TestAuthorizationHeader(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	if req.Header.Get("Authorization") == "" {
		t.Fatal("no authorization header")
	}
}

func runSigningTestsWithMaxAge(req *http.Request, maxAge time.Duration) bool {
	return ValidateAdminSigned(req, adminSecret, maxAge)
}

func runSigningTest(req *http.Request) bool {
	return runSigningTestsWithMaxAge(req, time.Minute)
}

func TestAdminSigningCorrectSignature(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	if !runSigningTest(req) {
		buf := bytes.Buffer{}
		concatenateTo(req, &buf)
		t.Log(buf.String())
		t.Fatal("validation failed")
	}
}

func TestAdminSigningEmptyAuthorizationHeader(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	req.Header.Set("Authorization", "")

	if runSigningTest(req) {
		t.Fatal("validation succeeded even without Authorization header")
	}
}

func TestAdminSigningNonLaraAuthorizationHeader(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	req.Header.Set("Authorization", "basic foo")

	if runSigningTest(req) {
		t.Fatal("validation succeeded even non-lara Authorization header")
	}
}

func TestAdminSigningNonAdminAuthorizationHeader(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	req.Header.Set("Authorization", "lara foo")
	if runSigningTest(req) {
		t.Fatal("validation succeeded even non-admin Authorization header")
	}
}

func TestAdminSigningNonGivenAuthorizationHeader(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	req.Header.Set("Authorization", "lara admin ")
	if runSigningTest(req) {
		t.Fatal("validation succeeded even without hash in Authorization header")
	}
}

func TestAdminSigningChangedMethodAuthorizationHeader(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	req.Method = "POST"
	if runSigningTest(req) {
		t.Fatal("validation succeeded even after method change")
	}
}

func TestAdminSigningAddedHeader(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	req.Header.Set("Test", "1")
	if runSigningTest(req) {
		t.Fatal("validation succeeded even after header addition")
	}
}

func TestAdminSigningChangedUrl(t *testing.T) {
	req := getRepositoriesAdminRequest(t)
	newURL, err := url.Parse("http://example.org/repositories?x=3")
	if err != nil {
		t.Fatal(err)
	}
	req.URL = newURL
	if runSigningTest(req) {
		t.Fatal("validation succeeded even after URL changing addition")
	}
}

func TestAdminSigningOutdatedSignature(t *testing.T) {
	req := getRepositoriesRequest(t)
	tenSecsAgo := time.Now().Add(-10 * time.Second)
	// this will most likely send non-GMT time which is against the HTTP RFC;
	// as we should handle this as well, it's ok for testing:
	req.Header.Set("Date", tenSecsAgo.Format(time.RFC1123))
	SignAsAdmin(req, adminSecret)
	if runSigningTestsWithMaxAge(req, 9*time.Second) {
		t.Fatal("validation succeeded even after reaching max signature age")
	}
}

func changedBodyAdminRequestForValidation(t *testing.T) *http.Request {
	req := getRepositoriesAdminRequest(t)
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("changed body")))
	return req
}

func TestAdminSigningBodyChange(t *testing.T) {
	req := changedBodyAdminRequestForValidation(t)
	if runSigningTest(req) {
		t.Fatal("validation succeeded even after body change")
	}
}

func TestAdminSigningBodyTextRead(t *testing.T) {
	req := changedBodyAdminRequestForValidation(t)

	runSigningTest(req)
	buf := make([]byte, 100)
	read, _ := req.Body.Read(buf)
	if string(buf[:read]) != "changed body" {
		t.Fatal("body no longer readable after signing")
	}
}
