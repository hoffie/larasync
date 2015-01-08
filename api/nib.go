package api

import (
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	repositoryModule "github.com/hoffie/larasync/repository"
)

// nibGet returns the NIB data for a given repository and a given UUID.
func (s *Server) nibGet(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repositoryName := vars["repository"]

	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		errorJSON(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	nibUUID := vars["nibUUID"]

	reader, err := repository.GetNIBReader(nibUUID)

	if err != nil {
		rw.Header().Set("Content-Type", "plain/text")
		if os.IsNotExist(err) {
			errorText(rw, "Not found", http.StatusNotFound)
		} else {
			errorText(rw, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	io.Copy(rw, reader)
}

// nibPut is the handler which adds a NIB to the repository.
func (s *Server) nibPut(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repositoryName := vars["repository"]

	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		errorJSON(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	nibUUID := vars["nibUUID"]

	err = repository.AddNIBContent(nibUUID, req.Body)

	if err != nil {
		if err == repositoryModule.SignatureVerificationError {
			errorText(rw, "Signature could not be verified", http.StatusBadRequest)
		} else if err == repositoryModule.UnMarshallingError {
			errorText(rw, "Could not extract NIB", http.StatusBadRequest)
		} else {
			errorText(rw, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	rw.Header().Set("Location", req.URL.String())
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) nibList(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repositoryName := vars["repository"]

	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		errorText(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	values := req.URL.Query()
	fromUUID, ok := values["from-uuid"]

	if !ok {
		errorText(rw, "from-uuid must be set", http.StatusBadRequest)
		return
	}

	nibChannel, err := repository.GetNIBsFrom(fromUUID[0])
	if err != nil {
		errorText(rw, "Could not extract data", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	for nib := range nibChannel {
		nib.WriteTo(rw)
	}
}
