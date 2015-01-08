package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

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
		if err == repositoryModule.ErrSignatureVerification {
			errorText(rw, "Signature could not be verified", http.StatusBadRequest)
		} else if err == repositoryModule.ErrUnMarshalling {
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
	fromRepositoryIDString, ok := values["from-transaction-id"]

	var nibChannel <-chan *repositoryModule.NIB
	if !ok {
		nibChannel, err = repository.GetAllNibs()
	} else {
		fromRepositoryID, err := strconv.ParseInt(fromRepositoryIDString[0], 10, 64)
		if err != nil {
			errorText(
				rw,
				fmt.Sprintf(
					"from-transaction-id %s is not a valid transaction-id",
					fromRepositoryIDString,
				),
				http.StatusBadRequest,
			)
			return
		}
		nibChannel, err = repository.GetNIBsFrom(fromRepositoryID)
	}

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
