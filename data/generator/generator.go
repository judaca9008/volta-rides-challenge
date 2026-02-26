package generator

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
	"voltarides/smart-router/models"

	"github.com/google/uuid"
)

// GenerateTestTransactions creates realistic test transaction data
func GenerateTestTransactions(count int) []models.Transaction {
	transactions := make([]models.Transaction, 0, count)
	now := time.Now()

	// Define processor configurations
	processorConfigs := []struct {
		name         string
		country      string
		currency     string
		approvalRate float64
		pattern      string // "normal", "strong", "bad_period"
	}{
		// Brazil
		{"RapidPay_BR", "BR", "BRL", 0.92, "strong"},      // 92% approval - strong performer
		{"TurboAcquire_BR", "BR", "BRL", 0.85, "normal"},  // 85% approval - normal
		{"PayFlow_BR", "BR", "BRL", 0.55, "bad_period"},   // 55-90% approval - has bad period

		// Mexico
		{"RapidPay_MX", "MX", "MXN", 0.90, "strong"},      // 90% approval - strong performer
		{"TurboAcquire_MX", "MX", "MXN", 0.82, "normal"},  // 82% approval - normal
		{"PayFlow_MX", "MX", "MXN", 0.56, "bad_period"},   // 56-88% approval - has bad period

		// Colombia
		{"RapidPay_CO", "CO", "COP", 0.91, "strong"},      // 91% approval - strong performer
		{"TurboAcquire_CO", "CO", "COP", 0.83, "normal"},  // 83% approval - normal
		{"PayFlow_CO", "CO", "COP", 0.57, "bad_period"},   // 57-89% approval - has bad period
	}

	// Generate transactions distributed across last 10 minutes (to fit within 15-minute time window)
	transactionsPerProcessor := count / len(processorConfigs)
	durationMinutes := 10.0 // Changed from 2.5 hours to 10 minutes to fit within 15-minute window
	intervalMinutes := durationMinutes / float64(transactionsPerProcessor)

	for _, cfg := range processorConfigs {
		for i := 0; i < transactionsPerProcessor; i++ {
			// Calculate timestamp (spread over last 2-3 hours)
			minutesAgo := int(float64(i) * intervalMinutes)
			timestamp := now.Add(-time.Duration(minutesAgo) * time.Minute)

			// Determine if this is during the "bad period" (middle portion of the time window)
			isBadPeriod := cfg.pattern == "bad_period" && minutesAgo >= 4 && minutesAgo <= 7

			// Calculate approval probability
			approvalProb := cfg.approvalRate
			if isBadPeriod {
				approvalProb = 0.55 // Drop to 55% during bad period
			}

			// Determine transaction status
			status := "approved"
			if rand.Float64() > approvalProb {
				status = "declined"
			}

			// Generate amount between $5 and $150
			amount := 5.0 + rand.Float64()*145.0

			transaction := models.Transaction{
				ID:        fmt.Sprintf("txn_%s", uuid.New().String()[:8]),
				Processor: cfg.name,
				Country:   cfg.country,
				Currency:  cfg.currency,
				Amount:    amount,
				Status:    status,
				Timestamp: timestamp,
			}

			transactions = append(transactions, transaction)
		}
	}

	return transactions
}

// SaveTransactionsToFile saves transactions to a JSON file
func SaveTransactionsToFile(transactions []models.Transaction, filepath string) error {
	dataset := models.TransactionDataset{
		Transactions: transactions,
	}

	data, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transactions: %w", err)
	}

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadTransactionsFromFile loads transactions from a JSON file
func LoadTransactionsFromFile(filepath string) ([]models.Transaction, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var dataset models.TransactionDataset
	err = json.Unmarshal(data, &dataset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return dataset.Transactions, nil
}
