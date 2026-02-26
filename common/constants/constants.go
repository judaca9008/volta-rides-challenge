package constants

const (
	// MicroserviceName is the name of this service
	MicroserviceName = "volta-router"

	// API version paths
	V1 = "/v1"

	// Route paths
	HealthCheck          = "/health"
	Route                = "/route"
	Processors           = "/processors"
	ProcessorByName      = "/processors/:name"
	RoutingStats         = "/routing/stats"
	TransactionsLoad     = "/transactions/load"
)
