package api

import (
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// blobGet is the handler to request a blob for a specific
// repository.
func (s *Server) blobGet(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repositoryName := vars["repository"]

	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		errorJSON(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	blobID := vars["blobID"]

	reader, err := repository.GetObjectData(blobID)

	if err != nil {
		if os.IsNotExist(err) {
			errorJSON(rw, "Not found", http.StatusNotFound)
		} else {
			errorJSON(rw, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	io.Copy(rw, reader)
}

// blobPut is the handler to set the content of a blob for a specific
// repository
func (s *Server) blobPut(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repositoryName := vars["repository"]

	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		errorJSON(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	blobID := vars["blobID"]

	err = repository.AddObject(blobID, req.Body)

	if err != nil {
		errorJSON(rw, "Internal Error", http.StatusInternalServerError)
	}

	rw.Header().Set("Location", req.URL.String())
	rw.WriteHeader(http.StatusOK)
}
