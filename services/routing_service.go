package services

import (
	"errors"
	"fmt"
	"time"
	"voltarides/smart-router/config"
	"voltarides/smart-router/models"
	"voltarides/smart-router/storage"
)

// RoutingService handles routing logic and approval rate calculations
type RoutingService struct {
	store  *storage.InMemoryStore
	config *config.RoutingConfig
}

// NewRoutingService creates a new routing service
func NewRoutingService(store *storage.InMemoryStore, cfg *config.RoutingConfig) *RoutingService {
	return &RoutingService{
		store:  store,
		config: cfg,
	}
}

// CalculateApprovalRate calculates the approval rate for a processor in a specific country
func (s *RoutingService) CalculateApprovalRate(processor, country string) float64 {
	transactions := s.store.GetTransactionsByWindow(processor, country, s.config.TimeWindow)

	if len(transactions) == 0 {
		return 0.0
	}

	approved := 0
	for _, tx := range transactions {
		if tx.IsApproved() {
			approved++
		}
	}

	return (float64(approved) / float64(len(transactions))) * 100.0
}

// SelectBestProcessor selects the best processor for a routing request
func (s *RoutingService) SelectBestProcessor(req models.RoutingRequest) (*models.RoutingResponse, error) {
	// Validate country
	processors, exists := config.ProcessorsByCountry[req.Country]
	if !exists {
		return nil, fmt.Errorf("country %s not supported", req.Country)
	}

	if len(processors) == 0 {
		return nil, errors.New("no processors available for country " + req.Country)
	}

	// Calculate approval rates for all processors in this country
	bestProcessor := ""
	bestRate := 0.0
	processorRates := make(map[string]float64)

	for _, processor := range processors {
		rate := s.CalculateApprovalRate(processor, req.Country)
		processorRates[processor] = rate

		if rate > bestRate {
			bestRate = rate
			bestProcessor = processor
		}
	}

	// If all processors have 0% rate, return error
	if bestRate == 0.0 {
		return nil, errors.New("no processor data available for country " + req.Country)
	}

	// Classify risk level
	riskLevel := s.classifyRiskLevel(bestRate)

	// Record the routing decision
	decision := models.RoutingDecision{
		Processor:    bestProcessor,
		Country:      req.Country,
		ApprovalRate: bestRate,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	s.store.RecordRoutingDecision(decision)

	// Build response
	reason := fmt.Sprintf("Highest approval rate for %s", req.Country)
	if riskLevel == "high" {
		reason = fmt.Sprintf("Best available processor for %s (all processors below 70%%)", req.Country)
	}

	return &models.RoutingResponse{
		Processor:    bestProcessor,
		ApprovalRate: bestRate,
		RiskLevel:    riskLevel,
		Reason:       reason,
		Timestamp:    time.Now().Format(time.RFC3339),
	}, nil
}

// classifyRiskLevel determines the risk level based on approval rate
func (s *RoutingService) classifyRiskLevel(approvalRate float64) string {
	if approvalRate < s.config.HighRiskThreshold {
		return "high"
	}
	if approvalRate < s.config.MediumRiskThreshold {
		return "medium"
	}
	return "low"
}

// GetAllProcessorStats returns stats for all processors across all countries
func (s *RoutingService) GetAllProcessorStats() []models.ProcessorStats {
	stats := make([]models.ProcessorStats, 0)

	for country, processors := range config.ProcessorsByCountry {
		for _, processor := range processors {
			stat := s.getProcessorStat(processor, country)
			stats = append(stats, stat)
		}
	}

	return stats
}

// GetProcessorStats returns stats for a specific processor
func (s *RoutingService) GetProcessorStats(name string) (*models.ProcessorStats, error) {
	// Find which country this processor belongs to
	for country, processors := range config.ProcessorsByCountry {
		for _, processor := range processors {
			if processor == name {
				stat := s.getProcessorStat(processor, country)
				return &stat, nil
			}
		}
	}

	return nil, fmt.Errorf("processor %s not found", name)
}

// getProcessorStat calculates stats for a single processor
func (s *RoutingService) getProcessorStat(processor, country string) models.ProcessorStats {
	transactions := s.store.GetTransactionsByWindow(processor, country, s.config.TimeWindow)
	approvalRate := s.CalculateApprovalRate(processor, country)

	return models.ProcessorStats{
		Name:             processor,
		Country:          country,
		ApprovalRate:     approvalRate,
		TransactionCount: len(transactions),
		LastUpdated:      time.Now().Format(time.RFC3339),
	}
}

// GetRoutingStats returns routing decision statistics
func (s *RoutingService) GetRoutingStats() models.RoutingStats {
	limit := 50 // Last 50 decisions
	distribution := s.store.GetRoutingStats(limit)
	totalDecisions := s.store.GetRoutingDecisionCount()

	return models.RoutingStats{
		TotalDecisions: totalDecisions,
		Distribution:   distribution,
		Window:         "last_50_decisions",
	}
}
