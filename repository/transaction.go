package repository

// Transaction represents a server side transaction for specific NIBs
// which is used to synchronize the different clients.
type Transaction struct {
	UUID         string
	NIBUUIDs     []string
	PreviousUUID string
}
