package api

import (
	"net/http"

	"encoding/json"
)

func errorJSON(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	error_obj := JSONError{
		Error: error,
	}
	data, _ := json.Marshal(error_obj)
	w.Write(data)
}
