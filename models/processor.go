package models

// ProcessorStats represents the health statistics for a processor
type ProcessorStats struct {
	Name             string  `json:"name"`
	Country          string  `json:"country"`
	ApprovalRate     float64 `json:"approval_rate"`
	TransactionCount int     `json:"transaction_count"`
	LastUpdated      string  `json:"last_updated"`
}

// ProcessorHealthResponse represents the response with all processor stats
type ProcessorHealthResponse struct {
	Processors []ProcessorStats `json:"processors"`
}

// HealthResponse represents a simple health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// LoadDataResponse represents the response after loading test data
type LoadDataResponse struct {
	Message          string `json:"message"`
	TransactionsLoaded int    `json:"transactions_loaded"`
}
