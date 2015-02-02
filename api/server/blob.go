package server

import (
    "fmt"
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
		errorText(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	blobID := vars["blobID"]

    Log.Debug(fmt.Sprintf("Repository %s: Requesting blob ID %s", repositoryName, blobID))
	reader, err := repository.GetObjectData(blobID)

	if err != nil {
		if os.IsNotExist(err) {
            Log.Debug(fmt.Sprintf("Repository %s: Could not find blob ID %s", repositoryName, blobID))
			errorText(rw, "Not found", http.StatusNotFound)
		} else {
            Log.Warn(fmt.Sprintf("Repository %s: Could not request blob ID %s", repositoryName, blobID))
			errorText(rw, "Internal Error", http.StatusInternalServerError)
		}
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	io.Copy(rw, reader)
	_ = reader.Close()
}

// blobPut is the handler to set the content of a blob for a specific
// repository
func (s *Server) blobPut(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repositoryName := vars["repository"]

	repository, err := s.rm.Open(repositoryName)
	if err != nil {
		errorText(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	blobID := vars["blobID"]

    Log.Debug(fmt.Sprintf("Repository: %s, Adding blob with ID %s", repositoryName, blobID))
	err = repository.AddObject(blobID, req.Body)

	if err != nil {
        Log.Warn(
            fmt.Sprintf(
                "Repository: %s, Could not add blob with ID %s. Error: %s",
                repositoryName, blobID, err.Error(),
            ),
        )
		errorText(rw, "Internal Error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Location", req.URL.String())
	rw.WriteHeader(http.StatusOK)
}
