package api

import (
	"net/http"
	"net/http/httptest"

	"testing"
	"time"
)

func adminSecret() []byte {
	return []byte("foo")
}

func getServer() *Server {
	return New(adminSecret(), time.Minute)
}

func getRepositoriesRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest("GET", "http://example.org/repositories", nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func getRepositoriesAdminRequest(t *testing.T) *http.Request {
	req := getRepositoriesRequest(t)
	SignAsAdmin(req, adminSecret())
	return req
}

func runTestRequestWith(server *Server, req *http.Request) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	server.router.ServeHTTP(rw, req)
	return rw
}

func runTestRequest(req *http.Request) *httptest.ResponseRecorder {
	return runTestRequestWith(getServer(), req)
}
