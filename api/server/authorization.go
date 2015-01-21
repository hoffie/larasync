package server

import (
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	. "github.com/hoffie/larasync/api/common"
)

// extractAuthorizationPubKey returns a public key which has been passed to the
// as the var "authorizationPublicKeyString" in the URL.
func extractAuthorizationPubKey(req *http.Request) (publicKey [PublicKeySize]byte) {
	vars := mux.Vars(req)

	publicKeyString := vars["authPublicKey"]
	publicKeySlice, err := hex.DecodeString(publicKeyString)
	publicKey = [PublicKeySize]byte{}
	if err != nil {
		publicKeySlice = []byte{}
	}
	copy(publicKey[:], publicKeySlice)
	return publicKey
}

// authorizationGet requests a authorization for a passed public key.
func (s *Server) authorizationGet(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	publicKey := extractAuthorizationPubKey(req)
	repositoryName := vars["repository"]
	repository, err := s.rm.Open(repositoryName)

	var readerErr error
	var reader io.ReadCloser

	if err == nil {
		// FIXME: This is a possible timing attack which exposes if the repository
		// does exist or not. At the moment there is no fix for this.
		reader, readerErr = repository.GetAuthorizationReader(publicKey)
	}

	if !ValidateRequest(req, publicKey, s.maxRequestAge) || err != nil || readerErr != nil {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)

	io.Copy(rw, reader)
	_ = reader.Close()

	repository.DeleteAuthorization(publicKey)
}

// authorizationPut adds a new authorization object to the repository.
func (s *Server) authorizationPut(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	publicKey := extractAuthorizationPubKey(req)
	repositoryName := vars["repository"]
	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
	}

	err = repository.SetAuthorizationData(publicKey, req.Body)
	if err != nil {
		http.Error(rw, "Internal Error", http.StatusInternalServerError)
	}

	rw.Header().Set("Location", req.URL.String())
	rw.WriteHeader(http.StatusCreated)
}
