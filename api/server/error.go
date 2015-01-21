package server

import (
	"net/http"

	"encoding/json"
	
	"github.com/hoffie/larasync/api"
)

func errorJSON(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorObj := api.JSONError{
		Error: error,
	}
	data, _ := json.Marshal(errorObj)
	w.Write(data)
}

func errorText(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(code)
	w.Write([]byte(error))
}
