package api

import (
	"encoding/json"
	"io/ioutil"

	"net/http"

	"github.com/gorilla/mux"
)

// repositoryList returns a list of all configured repositories.
func (s *Server) repositoryList(rw http.ResponseWriter, req *http.Request) {
	jsonHeader(rw)
	names, err := s.rm.ListNames()
	if err != nil {
		errorJson(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(names)
	if err != nil {
		errorJson(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	rw.Write(out)
}

func (s *Server) repositoryCreate(rw http.ResponseWriter, req *http.Request) {
	jsonHeader(rw)
	vars := mux.Vars(req)
	repository_name := vars["repository"]
	if s.rm.Exists(repository_name) {
		http.Error(rw, "Repository exists", http.StatusConflict)
		return
	}
	var repository JsonRepository
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &repository)
	if err != nil {
		http.Error(rw, "Bad Request", http.StatusBadRequest)
		return
	}

	err = s.rm.Create(repository_name, repository.pubKey)
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusCreated)

}
