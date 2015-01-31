package server

import (
	"net/http"

	"encoding/json"

	"github.com/hoffie/larasync/api"
)

func errorJSONMessage(w http.ResponseWriter, error string, code int) {
	errorObj := &api.JSONError{
		Error: error,
		Type:  "generic",
	}
	errorJSON(w, errorObj, code)
}

func errorJSON(w http.ResponseWriter, jsonError interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, _ := json.Marshal(jsonError)
	w.Write(data)
}

func errorText(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(code)
	w.Write([]byte(error))
}
