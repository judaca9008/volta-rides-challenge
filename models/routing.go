package models

// RoutingRequest represents a request to route a payment
type RoutingRequest struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Currency string  `json:"currency" validate:"required,len=3"`
	Country  string  `json:"country" validate:"required,len=2"`
}

// RoutingResponse represents the response with processor selection
type RoutingResponse struct {
	Processor    string  `json:"processor"`
	ApprovalRate float64 `json:"approval_rate"`
	RiskLevel    string  `json:"risk_level"` // "low", "medium", "high"
	Reason       string  `json:"reason"`
	Timestamp    string  `json:"timestamp"`
}

// RoutingDecision represents a historical routing decision for tracking
type RoutingDecision struct {
	Processor    string  `json:"processor"`
	Country      string  `json:"country"`
	ApprovalRate float64 `json:"approval_rate"`
	Timestamp    string  `json:"timestamp"`
}

// RoutingStats represents the routing statistics
type RoutingStats struct {
	TotalDecisions int            `json:"total_decisions"`
	Distribution   map[string]int `json:"distribution"`
	Window         string         `json:"window"`
}
