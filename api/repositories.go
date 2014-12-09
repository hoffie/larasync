package api

// JSONRepository structure which is being sent
// to the server when creating a new repository.
type JSONRepository struct {
	PubKey []byte `json:"pub_key"`
}
