package httpServer

import (
	"log"
	"voltarides/smart-router/config"
	"voltarides/smart-router/controllers"
	"voltarides/smart-router/routers"
	"voltarides/smart-router/services"
	"voltarides/smart-router/storage"

	"github.com/labstack/echo/v4"
)

// EchoServer represents the HTTP server
type EchoServer struct {
	Server *echo.Echo
}

// RunServer initializes and starts the Echo server
func (es *EchoServer) RunServer() {
	startServer := es.initServer()
	startServer()
}

// initServer initializes the Echo server with all dependencies
func (es *EchoServer) initServer() func() {
	// Initialize store
	store := storage.NewInMemoryStore()

	// Initialize config
	routingConfig := config.GetRoutingConfig()
	serverConfig := config.GetServerConfig()

	// Initialize service
	routingService := services.NewRoutingService(store, routingConfig)

	// Initialize controllers
	routingController := controllers.NewRoutingController(routingService)
	dataController := controllers.NewDataController(store)

	// Create Echo instance
	es.Server = echo.New()
	es.Server.HideBanner = true

	// Configure routes
	routers.ConfigRouter(es.Server, routingController, dataController)

	log.Printf("🚀 Volta Router initializing...")
	log.Printf("📍 Environment: %s", serverConfig.Environment)
	log.Printf("🔧 Time Window: %v", routingConfig.TimeWindow)
	log.Printf("⚠️  High Risk Threshold: %.1f%%", routingConfig.HighRiskThreshold)
	log.Printf("🌐 Server starting on port %s", serverConfig.Port)

	return func() {
		if err := es.Server.Start(":" + serverConfig.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}
