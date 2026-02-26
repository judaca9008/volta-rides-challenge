package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"voltarides/smart-router/data/generator"
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Generating test transaction data...")

	// Generate 540 transactions (60 per processor * 9 processors)
	transactions := generator.GenerateTestTransactions(540)

	// Save to file
	filepath := "data/test_transactions.json"
	err := generator.SaveTransactionsToFile(transactions, filepath)
	if err != nil {
		log.Fatalf("Failed to save transactions: %v", err)
	}

	fmt.Printf("✓ Generated %d transactions and saved to %s\n", len(transactions), filepath)
}
