package repository

type Transaction struct {
	UUID         string
	NIBUUIDs     []string
	PreviousUUID string
}
