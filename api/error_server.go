package api

import (
	"net/http"

	"encoding/json"
)

func errorJSON(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorObj := JSONError{
		Error: error,
	}
	data, _ := json.Marshal(errorObj)
	w.Write(data)
}
