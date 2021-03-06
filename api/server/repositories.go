package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"net/http"

	"github.com/gorilla/mux"

	"github.com/hoffie/larasync/api"
)

// repositoryList returns a list of all configured repositories.
func (s *Server) repositoryList(rw http.ResponseWriter, req *http.Request) {
	jsonHeader(rw)
	names, err := s.rm.ListNames()
	if err != nil {
		errorJSONMessage(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(names)
	if err != nil {
		errorJSONMessage(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	rw.Write(out)
}

func (s *Server) repositoryCreate(rw http.ResponseWriter, req *http.Request) {
	jsonHeader(rw)
	vars := mux.Vars(req)
	repositoryName := vars["repository"]
	if s.rm.Exists(repositoryName) {
		errorJSONMessage(rw, "Repository exists", http.StatusConflict)
		return
	}
	var repository api.JSONRepository
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorJSONMessage(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &repository)
	if err != nil {
		errorJSONMessage(rw, "Bad Request", http.StatusBadRequest)
		return
	}

	if len(repository.PubKey) != PublicKeySize {
		errorMessage := fmt.Sprintf(
			"Public key has to be of length %d got %d",
			PublicKeySize,
			len(repository.PubKey))
		errorJSONMessage(
			rw,
			errorMessage,
			http.StatusBadRequest)
	}

	err = s.rm.Create(repositoryName, repository.PubKey)
	if err != nil {
		errorJSONMessage(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusCreated)

}
