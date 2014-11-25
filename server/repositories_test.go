package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hoffie/lara/sign/request"
)

func TestRepoList(t *testing.T) {
	testAdminSecret := []byte("foo")
	s := New(testAdminSecret)

	req, err := http.NewRequest("GET", "http://example.org/repositories", nil)
	if err != nil {
		t.Fatal(err)
	}

	rw := httptest.NewRecorder()

	s.router.ServeHTTP(rw, req)
	if rw.Code != 401 {
		t.Fatal("expected unauthorized but got", rw.Code)
	}
	if rw.Body.String() != "Unauthorized\n" {
		t.Fatal("unexpected body")
	}

	request.SignAsAdmin(req, testAdminSecret)
	rw = httptest.NewRecorder()

	s.router.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Fatal("expected HTTP 200 but got", rw.Code)
	}

	//FIXME test repo list output
}
