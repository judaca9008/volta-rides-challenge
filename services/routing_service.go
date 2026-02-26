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
func (s *RoutingService) SelectBestProcessor(req models.RoutingRequest, simulate bool) (*models.RoutingResponse, error) {
	return s.selectProcessor(req, simulate, false)
}

// SelectBestProcessorWithFailover selects the best processor and provides failover options
func (s *RoutingService) SelectBestProcessorWithFailover(req models.RoutingRequest, simulate bool) (*models.RoutingResponse, error) {
	return s.selectProcessor(req, simulate, true)
}

// selectProcessor is the internal implementation for processor selection
func (s *RoutingService) selectProcessor(req models.RoutingRequest, simulate bool, includeFailover bool) (*models.RoutingResponse, error) {
	// Validate country
	processors, exists := config.ProcessorsByCountry[req.Country]
	if !exists {
		return nil, fmt.Errorf("country %s not supported", req.Country)
	}

	if len(processors) == 0 {
		return nil, errors.New("no processors available for country " + req.Country)
	}

	// Calculate approval rates for all processors in this country
	type processorRate struct {
		name string
		rate float64
	}

	rates := make([]processorRate, 0, len(processors))
	for _, processor := range processors {
		// Check circuit breaker state
		circuitState := s.store.GetCircuitState(processor, req.Country, s.config.CircuitBreakerTimeout)

		// Skip processors with open circuit breaker
		if circuitState == models.CircuitOpen {
			continue
		}

		rate := s.CalculateApprovalRate(processor, req.Country)

		// Check if circuit should be opened
		if rate > 0 && rate < s.config.CircuitBreakerThreshold {
			s.store.OpenCircuit(processor, req.Country)
			continue // Skip this processor
		}

		// If circuit is half-open and rate is good, close it
		if circuitState == models.CircuitHalfOpen && rate >= s.config.CircuitBreakerThreshold {
			s.store.CloseCircuit(processor, req.Country)
		}

		rates = append(rates, processorRate{name: processor, rate: rate})
	}

	// Sort processors by approval rate (descending)
	for i := 0; i < len(rates); i++ {
		for j := i + 1; j < len(rates); j++ {
			if rates[j].rate > rates[i].rate {
				rates[i], rates[j] = rates[j], rates[i]
			}
		}
	}

	// If all processors have 0% rate or all circuits are open, return error
	if len(rates) == 0 || rates[0].rate == 0.0 {
		return nil, errors.New("no processor data available for country " + req.Country)
	}

	bestProcessor := rates[0].name
	bestRate := rates[0].rate

	// Classify risk level
	riskLevel := s.classifyRiskLevel(bestRate)

	// Record the routing decision (only if not in simulation mode)
	if !simulate {
		decision := models.RoutingDecision{
			Processor:    bestProcessor,
			Country:      req.Country,
			ApprovalRate: bestRate,
			Timestamp:    time.Now().Format(time.RFC3339),
		}
		s.store.RecordRoutingDecision(decision)
	}

	// Build response
	reason := fmt.Sprintf("Highest approval rate for %s", req.Country)
	if riskLevel == "high" {
		reason = fmt.Sprintf("Best available processor for %s (all processors below 70%%)", req.Country)
	}

	response := &models.RoutingResponse{
		Processor:    bestProcessor,
		ApprovalRate: bestRate,
		RiskLevel:    riskLevel,
		Reason:       reason,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	// Add failover options if requested and available
	if includeFailover {
		if len(rates) > 1 && rates[1].rate > 0 {
			response.Fallback = &models.ProcessorOption{
				Processor:    rates[1].name,
				ApprovalRate: rates[1].rate,
			}
		}
		if len(rates) > 2 && rates[2].rate > 0 {
			response.LastResort = &models.ProcessorOption{
				Processor:    rates[2].name,
				ApprovalRate: rates[2].rate,
			}
		}
	}

	return response, nil
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

	// Get circuit breaker state
	circuitState := s.store.GetCircuitState(processor, country, s.config.CircuitBreakerTimeout)

	stat := models.ProcessorStats{
		Name:             processor,
		Country:          country,
		ApprovalRate:     approvalRate,
		TransactionCount: len(transactions),
		LastUpdated:      time.Now().Format(time.RFC3339),
	}

	// Add circuit breaker info if not closed
	if circuitState != models.CircuitClosed {
		stat.CircuitState = circuitState
		if openedAt := s.store.GetCircuitOpenedAt(processor, country); openedAt != nil {
			stat.CircuitOpenedAt = openedAt.Format(time.RFC3339)
		}
	}

	return stat
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
