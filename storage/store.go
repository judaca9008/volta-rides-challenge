package storage

import (
	"sync"
	"time"
	"voltarides/smart-router/models"
)

// InMemoryStore provides thread-safe in-memory storage for transactions and routing decisions
type InMemoryStore struct {
	transactions     []models.Transaction
	routingDecisions []models.RoutingDecision
	mu               sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		transactions:     make([]models.Transaction, 0),
		routingDecisions: make([]models.RoutingDecision, 0),
	}
}

// AddTransaction adds a transaction to the store
func (s *InMemoryStore) AddTransaction(tx models.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactions = append(s.transactions, tx)
}

// AddTransactions adds multiple transactions to the store
func (s *InMemoryStore) AddTransactions(txs []models.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactions = append(s.transactions, txs...)
}

// GetTransactionsByWindow returns transactions for a specific processor and country within a time window
func (s *InMemoryStore) GetTransactionsByWindow(processor, country string, window time.Duration) []models.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoffTime := time.Now().Add(-window)
	filtered := make([]models.Transaction, 0)

	for _, tx := range s.transactions {
		if tx.Processor == processor && tx.Country == country && tx.Timestamp.After(cutoffTime) {
			filtered = append(filtered, tx)
		}
	}

	return filtered
}

// GetAllTransactions returns all transactions (thread-safe copy)
func (s *InMemoryStore) GetAllTransactions() []models.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]models.Transaction, len(s.transactions))
	copy(result, s.transactions)
	return result
}

// GetTransactionCount returns the total number of transactions
func (s *InMemoryStore) GetTransactionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.transactions)
}

// RecordRoutingDecision records a routing decision for tracking
func (s *InMemoryStore) RecordRoutingDecision(decision models.RoutingDecision) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routingDecisions = append(s.routingDecisions, decision)
}

// GetRoutingStats returns the routing decision distribution for the last N decisions
func (s *InMemoryStore) GetRoutingStats(limit int) map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	distribution := make(map[string]int)
	start := 0
	if len(s.routingDecisions) > limit {
		start = len(s.routingDecisions) - limit
	}

	for i := start; i < len(s.routingDecisions); i++ {
		processor := s.routingDecisions[i].Processor
		distribution[processor]++
	}

	return distribution
}

// GetRoutingDecisionCount returns the total number of routing decisions
func (s *InMemoryStore) GetRoutingDecisionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.routingDecisions)
}

// Clear removes all data from the store (useful for testing)
func (s *InMemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactions = make([]models.Transaction, 0)
	s.routingDecisions = make([]models.RoutingDecision, 0)
}
