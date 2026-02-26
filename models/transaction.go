package models

import "time"

// Transaction represents a payment transaction
type Transaction struct {
	ID        string    `json:"id"`
	Processor string    `json:"processor"`
	Country   string    `json:"country"`
	Currency  string    `json:"currency"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"` // "approved" or "declined"
	Timestamp time.Time `json:"timestamp"`
}

// IsApproved returns true if the transaction was approved
func (t *Transaction) IsApproved() bool {
	return t.Status == "approved"
}

// TransactionDataset represents the JSON structure for loading test data
type TransactionDataset struct {
	Transactions []Transaction `json:"transactions"`
}
