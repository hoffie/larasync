package api

// JSONError is the structure which gets send
// to the client when a JSON endpoint encounters
// an error
type JSONError struct {
	Type  string `json:"error_type"`
	Error string `json:"error"`
}

// ContentIDsJSONError gets returned if there is content
// being referenced in an uploaded NIB is missing.
type ContentIDsJSONError struct {
	Type              string   `json:"error_type"`
	Error             string   `json:"error"`
	MissingContentIDs []string `json:"missing_content_ids"`
}
