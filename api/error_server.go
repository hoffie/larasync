package api

import (
	"net/http"

	"encoding/json"
)

func errorJson(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	error_obj := JsonError{
		error: error,
	}
	data, _ := json.Marshal(error_obj)
	w.Write(data)
}
