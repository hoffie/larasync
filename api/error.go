package api

// JSONError is the structure which gets send
// to the client when a JSON endpoint encounters
// an error
type JSONError struct {
	Error string `json:"error"`
}
