package constants

// TransactionType represents the type of transaction (buy or sell)
type TransactionType int

const (
	BUY TransactionType = iota
	SELL
)
