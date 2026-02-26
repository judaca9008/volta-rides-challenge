package config

import (
	"os"
	"time"
)

// RoutingConfig holds configuration for the routing service
type RoutingConfig struct {
	TimeWindow          time.Duration
	HighRiskThreshold   float64
	MediumRiskThreshold float64
	CircuitBreakerThreshold float64
	CircuitBreakerTimeout   time.Duration
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	Environment string
}

// GetRoutingConfig returns the routing configuration with defaults
func GetRoutingConfig() *RoutingConfig {
	return &RoutingConfig{
		TimeWindow:          15 * time.Minute,  // Default: last 15 minutes
		HighRiskThreshold:   70.0,              // Below 70% is high risk
		MediumRiskThreshold: 80.0,              // 70-80% is medium risk
		CircuitBreakerThreshold: 60.0,          // Below 60% opens circuit
		CircuitBreakerTimeout:   5 * time.Minute, // Circuit stays open for 5 min
	}
}

// GetServerConfig returns the server configuration from environment variables
func GetServerConfig() *ServerConfig {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	return &ServerConfig{
		Port:        port,
		Environment: environment,
	}
}

// ProcessorsByCountry defines the mapping of countries to their processors
var ProcessorsByCountry = map[string][]string{
	"BR": {"RapidPay_BR", "TurboAcquire_BR", "PayFlow_BR"},
	"MX": {"RapidPay_MX", "TurboAcquire_MX", "PayFlow_MX"},
	"CO": {"RapidPay_CO", "TurboAcquire_CO", "PayFlow_CO"},
}
