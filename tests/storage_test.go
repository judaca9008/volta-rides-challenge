package tests

import (
	"sync"
	"testing"
	"time"
	"voltarides/smart-router/models"
	"voltarides/smart-router/storage"
)

func TestConcurrentAddTransactions(t *testing.T) {
	store := storage.NewInMemoryStore()
	now := time.Now()

	// Number of concurrent goroutines
	numGoroutines := 100
	transactionsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < transactionsPerGoroutine; j++ {
				tx := models.Transaction{
					ID:        "tx_concurrent",
					Processor: "RapidPay_BR",
					Country:   "BR",
					Status:    "approved",
					Timestamp: now,
				}
				store.AddTransaction(tx)
			}
		}(i)
	}

	wg.Wait()

	// Verify total count
	expectedCount := numGoroutines * transactionsPerGoroutine
	actualCount := store.GetTransactionCount()

	if actualCount != expectedCount {
		t.Errorf("Expected %d transactions, got %d (possible race condition)", expectedCount, actualCount)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	store := storage.NewInMemoryStore()
	now := time.Now()

	// Add initial transactions
	for i := 0; i < 50; i++ {
		tx := models.Transaction{
			ID:        "tx_initial",
			Processor: "RapidPay_BR",
			Country:   "BR",
			Status:    "approved",
			Timestamp: now,
		}
		store.AddTransaction(tx)
	}

	var wg sync.WaitGroup
	numReaders := 50
	numWriters := 10

	// Launch concurrent readers
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			// Read operations
			for j := 0; j < 100; j++ {
				_ = store.GetTransactionsByWindow("RapidPay_BR", "BR", 15*time.Minute)
				_ = store.GetTransactionCount()
				_ = store.GetAllTransactions()
			}
		}()
	}

	// Launch concurrent writers
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func() {
			defer wg.Done()
			// Write operations
			for j := 0; j < 10; j++ {
				tx := models.Transaction{
					ID:        "tx_writer",
					Processor: "RapidPay_BR",
					Country:   "BR",
					Status:    "approved",
					Timestamp: now,
				}
				store.AddTransaction(tx)
			}
		}()
	}

	wg.Wait()

	// If we reach here without deadlock or panic, the test passes
	t.Log("Concurrent read/write operations completed successfully")
}

func TestGetTransactionsByWindow(t *testing.T) {
	store := storage.NewInMemoryStore()
	now := time.Now()

	// Add transactions with different timestamps
	transactions := []models.Transaction{
		{ID: "tx1", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},  // Within 15min
		{ID: "tx2", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-10 * time.Minute)}, // Within 15min
		{ID: "tx3", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-20 * time.Minute)}, // Outside 15min
		{ID: "tx4", Processor: "RapidPay_MX", Country: "MX", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},  // Different country
	}
	store.AddTransactions(transactions)

	// Query for BR transactions within 15 minutes
	results := store.GetTransactionsByWindow("RapidPay_BR", "BR", 15*time.Minute)

	// Should only get 2 transactions (tx1, tx2)
	if len(results) != 2 {
		t.Errorf("Expected 2 transactions within window, got %d", len(results))
	}

	// Verify IDs
	for _, tx := range results {
		if tx.ID != "tx1" && tx.ID != "tx2" {
			t.Errorf("Unexpected transaction ID in results: %s", tx.ID)
		}
	}
}

func TestRoutingDecisionTracking(t *testing.T) {
	store := storage.NewInMemoryStore()

	// Record routing decisions
	decisions := []models.RoutingDecision{
		{Processor: "RapidPay_BR", Country: "BR", ApprovalRate: 92.5, Timestamp: time.Now().Format(time.RFC3339)},
		{Processor: "RapidPay_BR", Country: "BR", ApprovalRate: 91.0, Timestamp: time.Now().Format(time.RFC3339)},
		{Processor: "TurboAcquire_BR", Country: "BR", ApprovalRate: 85.0, Timestamp: time.Now().Format(time.RFC3339)},
		{Processor: "RapidPay_BR", Country: "BR", ApprovalRate: 93.0, Timestamp: time.Now().Format(time.RFC3339)},
	}

	for _, decision := range decisions {
		store.RecordRoutingDecision(decision)
	}

	// Get stats for last 10 decisions
	stats := store.GetRoutingStats(10)

	// Verify distribution
	if stats["RapidPay_BR"] != 3 {
		t.Errorf("Expected 3 decisions for RapidPay_BR, got %d", stats["RapidPay_BR"])
	}

	if stats["TurboAcquire_BR"] != 1 {
		t.Errorf("Expected 1 decision for TurboAcquire_BR, got %d", stats["TurboAcquire_BR"])
	}

	// Verify total count
	totalCount := store.GetRoutingDecisionCount()
	if totalCount != 4 {
		t.Errorf("Expected 4 total decisions, got %d", totalCount)
	}
}

func TestStoreClear(t *testing.T) {
	store := storage.NewInMemoryStore()
	now := time.Now()

	// Add some data
	store.AddTransaction(models.Transaction{
		ID:        "tx1",
		Processor: "RapidPay_BR",
		Country:   "BR",
		Status:    "approved",
		Timestamp: now,
	})

	store.RecordRoutingDecision(models.RoutingDecision{
		Processor:    "RapidPay_BR",
		Country:      "BR",
		ApprovalRate: 92.5,
		Timestamp:    now.Format(time.RFC3339),
	})

	// Verify data exists
	if store.GetTransactionCount() == 0 {
		t.Fatal("Expected transactions to be added")
	}
	if store.GetRoutingDecisionCount() == 0 {
		t.Fatal("Expected routing decisions to be added")
	}

	// Clear store
	store.Clear()

	// Verify store is empty
	if store.GetTransactionCount() != 0 {
		t.Errorf("Expected 0 transactions after clear, got %d", store.GetTransactionCount())
	}
	if store.GetRoutingDecisionCount() != 0 {
		t.Errorf("Expected 0 routing decisions after clear, got %d", store.GetRoutingDecisionCount())
	}
}

func TestConcurrentRoutingDecisions(t *testing.T) {
	store := storage.NewInMemoryStore()

	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent routing decision writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			decision := models.RoutingDecision{
				Processor:    "RapidPay_BR",
				Country:      "BR",
				ApprovalRate: 92.5,
				Timestamp:    time.Now().Format(time.RFC3339),
			}
			store.RecordRoutingDecision(decision)
		}(i)
	}

	wg.Wait()

	// Verify count
	count := store.GetRoutingDecisionCount()
	if count != numGoroutines {
		t.Errorf("Expected %d routing decisions, got %d (possible race condition)", numGoroutines, count)
	}
}
