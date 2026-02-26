package tests

import (
	"testing"
	"time"
	"voltarides/smart-router/config"
	"voltarides/smart-router/models"
	"voltarides/smart-router/services"
	"voltarides/smart-router/storage"
)

func TestCalculateApprovalRate(t *testing.T) {
	store := storage.NewInMemoryStore()
	cfg := config.GetRoutingConfig()
	service := services.NewRoutingService(store, cfg)

	// Add test transactions
	now := time.Now()
	transactions := []models.Transaction{
		{ID: "tx1", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx2", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx3", Processor: "RapidPay_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx4", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx5", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
	}
	store.AddTransactions(transactions)

	// Calculate approval rate: 4 approved out of 5 = 80%
	rate := service.CalculateApprovalRate("RapidPay_BR", "BR")

	if rate != 80.0 {
		t.Errorf("Expected approval rate 80.0, got %.2f", rate)
	}
}

func TestCalculateApprovalRateNoData(t *testing.T) {
	store := storage.NewInMemoryStore()
	cfg := config.GetRoutingConfig()
	service := services.NewRoutingService(store, cfg)

	// No transactions
	rate := service.CalculateApprovalRate("RapidPay_BR", "BR")

	if rate != 0.0 {
		t.Errorf("Expected approval rate 0.0 for no data, got %.2f", rate)
	}
}

func TestCalculateApprovalRateTimeWindow(t *testing.T) {
	store := storage.NewInMemoryStore()
	cfg := config.GetRoutingConfig()
	service := services.NewRoutingService(store, cfg)

	now := time.Now()

	// Add transactions: 2 within window, 1 outside window
	transactions := []models.Transaction{
		{ID: "tx1", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},  // Within
		{ID: "tx2", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-10 * time.Minute)}, // Within
		{ID: "tx3", Processor: "RapidPay_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-20 * time.Minute)}, // Outside 15min window
	}
	store.AddTransactions(transactions)

	// Should only count the 2 transactions within 15-minute window
	rate := service.CalculateApprovalRate("RapidPay_BR", "BR")

	if rate != 100.0 {
		t.Errorf("Expected approval rate 100.0 (only counting transactions within window), got %.2f", rate)
	}
}

func TestSelectBestProcessor(t *testing.T) {
	store := storage.NewInMemoryStore()
	cfg := config.GetRoutingConfig()
	service := services.NewRoutingService(store, cfg)

	now := time.Now()

	// Add transactions for multiple processors in Brazil
	transactions := []models.Transaction{
		// RapidPay_BR: 90% approval (9/10)
		{ID: "tx1", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx2", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx3", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx4", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx5", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx6", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx7", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx8", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx9", Processor: "RapidPay_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx10", Processor: "RapidPay_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},

		// TurboAcquire_BR: 70% approval (7/10)
		{ID: "tx11", Processor: "TurboAcquire_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx12", Processor: "TurboAcquire_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx13", Processor: "TurboAcquire_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx14", Processor: "TurboAcquire_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx15", Processor: "TurboAcquire_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx16", Processor: "TurboAcquire_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx17", Processor: "TurboAcquire_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx18", Processor: "TurboAcquire_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx19", Processor: "TurboAcquire_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx20", Processor: "TurboAcquire_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},

		// PayFlow_BR: 50% approval (5/10)
		{ID: "tx21", Processor: "PayFlow_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx22", Processor: "PayFlow_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx23", Processor: "PayFlow_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx24", Processor: "PayFlow_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx25", Processor: "PayFlow_BR", Country: "BR", Status: "approved", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx26", Processor: "PayFlow_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx27", Processor: "PayFlow_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx28", Processor: "PayFlow_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx29", Processor: "PayFlow_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},
		{ID: "tx30", Processor: "PayFlow_BR", Country: "BR", Status: "declined", Timestamp: now.Add(-5 * time.Minute)},
	}
	store.AddTransactions(transactions)

	// Create routing request
	req := models.RoutingRequest{
		Amount:   100.0,
		Currency: "BRL",
		Country:  "BR",
	}

	response, err := service.SelectBestProcessor(req, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should select RapidPay_BR (90% approval rate)
	if response.Processor != "RapidPay_BR" {
		t.Errorf("Expected processor RapidPay_BR, got %s", response.Processor)
	}

	if response.ApprovalRate != 90.0 {
		t.Errorf("Expected approval rate 90.0, got %.2f", response.ApprovalRate)
	}

	if response.RiskLevel != "low" {
		t.Errorf("Expected risk level 'low', got %s", response.RiskLevel)
	}
}

func TestSelectBestProcessorUnsupportedCountry(t *testing.T) {
	store := storage.NewInMemoryStore()
	cfg := config.GetRoutingConfig()
	service := services.NewRoutingService(store, cfg)

	req := models.RoutingRequest{
		Amount:   100.0,
		Currency: "USD",
		Country:  "US",
	}

	_, err := service.SelectBestProcessor(req, false)
	if err == nil {
		t.Fatal("Expected error for unsupported country, got nil")
	}

	expectedError := "country US not supported"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSelectBestProcessorNoData(t *testing.T) {
	store := storage.NewInMemoryStore()
	cfg := config.GetRoutingConfig()
	service := services.NewRoutingService(store, cfg)

	// No transactions in store
	req := models.RoutingRequest{
		Amount:   100.0,
		Currency: "BRL",
		Country:  "BR",
	}

	_, err := service.SelectBestProcessor(req, false)
	if err == nil {
		t.Fatal("Expected error when no data available, got nil")
	}

	expectedError := "no processor data available for country BR"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestRiskLevelClassification(t *testing.T) {
	store := storage.NewInMemoryStore()
	cfg := config.GetRoutingConfig()
	service := services.NewRoutingService(store, cfg)

	now := time.Now()

	testCases := []struct {
		name              string
		approvedCount     int
		totalCount        int
		expectedRiskLevel string
	}{
		{"Low Risk", 9, 10, "low"},     // 90% > 80%
		{"Medium Risk", 75, 100, "medium"}, // 75% between 70-80%
		{"High Risk", 6, 10, "high"},   // 60% < 70%
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear store
			store.Clear()

			// Add transactions
			transactions := make([]models.Transaction, 0, tc.totalCount)
			for i := 0; i < tc.totalCount; i++ {
				status := "declined"
				if i < tc.approvedCount {
					status = "approved"
				}
				transactions = append(transactions, models.Transaction{
					ID:        "tx_" + string(rune(i)),
					Processor: "TestProcessor",
					Country:   "BR",
					Status:    status,
					Timestamp: now.Add(-5 * time.Minute),
				})
			}
			store.AddTransactions(transactions)

			// Test routing
			req := models.RoutingRequest{
				Amount:   100.0,
				Currency: "BRL",
				Country:  "BR",
			}

			// Temporarily modify config to include TestProcessor
			originalProcessors := config.ProcessorsByCountry["BR"]
			config.ProcessorsByCountry["BR"] = []string{"TestProcessor"}
			defer func() {
				config.ProcessorsByCountry["BR"] = originalProcessors
			}()

			response, err := service.SelectBestProcessor(req, false)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if response.RiskLevel != tc.expectedRiskLevel {
				t.Errorf("Expected risk level '%s', got '%s'", tc.expectedRiskLevel, response.RiskLevel)
			}
		})
	}
}
