package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/hoffie/larasync/api"
	"github.com/hoffie/larasync/helpers/bincontainer"
	repositoryModule "github.com/hoffie/larasync/repository"
)

// nibGet returns the NIB data for a given repository and a given UUID.
func (s *Server) nibGet(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repositoryName := vars["repository"]

	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		errorJSONMessage(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	nibID := vars["nibID"]

	Log.Debug(fmt.Sprintf("Repository %s: Requesting NIB with ID %s", repositoryName, nibID))
	reader, err := repository.GetNIBReader(nibID)

	if err != nil {
		rw.Header().Set("Content-Type", "plain/text")
		if os.IsNotExist(err) {
			errorText(rw, "Not found", http.StatusNotFound)
		} else {
			errorText(rw, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	defer reader.Close()

	rw.Header().Set("Content-Type", "application/octet-stream")
	attachCurrentTransactionHeader(repository, rw)
	rw.WriteHeader(http.StatusOK)
	io.Copy(rw, reader)
}

// nibPut is the handler which adds a NIB to the repository.
func (s *Server) nibPut(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repositoryName := vars["repository"]

	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		errorJSONMessage(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	nibID := vars["nibID"]

	successReturnStatus := http.StatusOK
	if !repository.HasNIB(nibID) {
		successReturnStatus = http.StatusCreated
	}

	Log.Debug(fmt.Sprintf("Repository %s: Adding NIB with ID %s", repositoryName, nibID))
	err = repository.AddNIBContent(req.Body)

	if err != nil {
		if err == repositoryModule.ErrSignatureVerification {
			Log.Debug(fmt.Sprintf("Repository %s: Signature Verification failed when trying to add NIB with ID %s", repositoryName, nibID))
			errorText(rw, "Signature could not be verified", http.StatusUnauthorized)
		} else if err == repositoryModule.ErrUnMarshalling {
			Log.Debug(fmt.Sprintf("Repository %s: Unmarshalling failed when trying to add NIB with ID %s", repositoryName, nibID))
			errorText(rw, "Could not extract NIB", http.StatusBadRequest)
		} else if err == repositoryModule.ErrNIBConflict {
			Log.Debug(fmt.Sprintf("Repository %s: Conflict when trying to add NIB with ID %s", repositoryName, nibID))
			errorText(rw, "NIB conflict", http.StatusConflict)
		} else if repositoryModule.IsNIBContentMissing(err) {
			Log.Debug(fmt.Sprintf("Repository %s: Contents of NIB not in Server when trying to add NIB with ID %s", repositoryName, nibID))
			nibError := err.(*repositoryModule.ErrNIBContentMissing)
			jsonError := &api.ContentIDsJSONError{}
			jsonError.Error = nibError.Error()
			jsonError.Type = "missing_content_ids"
			jsonError.MissingContentIDs = nibError.MissingContentIDs()
			errorJSON(rw, jsonError, http.StatusPreconditionFailed)
		} else {
			Log.Warn(fmt.Sprintf("Repository %s: Internal Error when trying to add NIB with ID %s. %s", repositoryName, nibID, err.Error()))
			errorText(rw, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	rw.Header().Set("Location", req.URL.String())
	attachCurrentTransactionHeader(repository, rw)
	rw.WriteHeader(successReturnStatus)
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

	var nibChannel <-chan []byte
	if !ok {
		Log.Debug(fmt.Sprintf("Repository %s:, Requesting complete NIB list", repositoryName))
		nibChannel, err = repository.GetAllNIBBytes()
	} else {
		afterTransactionID, err := strconv.ParseInt(fromRepositoryIDString[0], 10, 64)
		if err != nil {
			errorText(
				rw,
				fmt.Sprintf(
					"from-transaction-id %s is not a valid transaction id",
					fromRepositoryIDString,
				),
				http.StatusBadRequest,
			)
			Log.Debug(fmt.Sprintf("Repository %s: Error while trying to extract transaction id of %s", repositoryName, fromRepositoryIDString[0]))
			return
		}
		Log.Debug(fmt.Sprintf("Repository %s: Requesting NIB list after transaction id %d", repositoryName, afterTransactionID))
		nibChannel, err = repository.GetNIBBytesFrom(afterTransactionID)
	}

	if err != nil {
		Log.Warn(fmt.Sprintf("Repository %s: Could not extract nib data.", repositoryName))
		errorText(rw, "Could not extract data", http.StatusInternalServerError)
		return
	}

	header := rw.Header()
	header.Set("Content-Type", "application/octet-stream")
	attachCurrentTransactionHeader(repository, rw)

	rw.WriteHeader(http.StatusOK)

	encoder := bincontainer.NewEncoder(rw)
	for nibData := range nibChannel {
		encoder.WriteChunk(nibData)
	}
}
