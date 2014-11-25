package api

import (
	"testing"
)

func TestRepoListUnauthorized(t *testing.T) {
	resp := runTestRequest(
		getRepositoriesRequest(t))
	if resp.Code != 401 {
		t.Fatal("expected unauthorized but got", resp.Code)
	}
	if resp.Body.String() != "Unauthorized\n" {
		t.Fatal("unexpected unauthorized body")
	}
}

func TestRepoListAdmin(t *testing.T) {
	req := getRepositoriesAdminRequest(t)

	resp := runTestRequest(req)
	if resp.Code != 200 {
		t.Fatal("expected HTTP 200 but got", resp.Code)
	}
}

func TestRepoListOutput(t *testing.T) {
	req := getRepositoriesAdminRequest(t)

	resp := runTestRequest(req)

	//FIXME test repo list output
	if resp.Body.Len() == 0 {
		t.Fatal("expected HTTP body not to be empty.")
	}
}
