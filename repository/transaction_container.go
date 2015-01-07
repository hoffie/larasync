package repository

// TransactionContainer is used to encapsulate several transactions
// in one data format.
type TransactionContainer struct {
	UUID         string
	Transactions []*Transaction
	PreviousUUID string
}
