package controllers

import (
	"net/http"
	"voltarides/smart-router/data/generator"
	"voltarides/smart-router/models"
	"voltarides/smart-router/storage"

	"github.com/labstack/echo/v4"
)

// DataController handles test data operations
type DataController struct {
	store *storage.InMemoryStore
}

// NewDataController creates a new data controller
func NewDataController(store *storage.InMemoryStore) *DataController {
	return &DataController{store: store}
}

// LoadTestData loads test transactions from the JSON file
func (dc *DataController) LoadTestData(c echo.Context) error {
	filepath := "data/test_transactions.json"

	// Load transactions from file
	transactions, err := generator.LoadTransactionsFromFile(filepath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "load_failed",
			Message: "Failed to load test data: " + err.Error(),
		})
	}

	// Clear existing data
	dc.store.Clear()

	// Add transactions to store
	dc.store.AddTransactions(transactions)

	return c.JSON(http.StatusOK, models.LoadDataResponse{
		Message:            "Test data loaded successfully",
		TransactionsLoaded: len(transactions),
	})
}
