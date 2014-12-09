package api

import (
	"encoding/json"

	"net/http"
)

// repositoryList returns a list of all configured repositories.
func (s *Server) repositoryList(rw http.ResponseWriter, req *http.Request) {
	jsonHeader(rw)
	names, err := s.rm.ListNames()
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(names)
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	rw.Write(out)
}
