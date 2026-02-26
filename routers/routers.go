package routers

import (
	"voltarides/smart-router/common/constants"
	"voltarides/smart-router/controllers"
	middlewareLog "voltarides/smart-router/routers/middleware/log"
	"voltarides/smart-router/routers/middleware/trace_id"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoDatadog "gopkg.in/DataDog/dd-trace-go.v1/contrib/labstack/echo.v4"
)

// ConfigRouter configures all routes and middleware for the Echo server
func ConfigRouter(
	e *echo.Echo,
	routingController *controllers.RoutingController,
	dataController *controllers.DataController,
) {
	// Middleware stack (Yuno standard pattern)
	// 1. DataDog APM (distributed tracing)
	e.Use(echoDatadog.Middleware(echoDatadog.WithServiceName(constants.MicroserviceName)))

	// 2. Logging middleware
	e.Use(middlewareLog.LoggingMiddleware())

	// 3. Trace ID propagation
	e.Use(trace_id.TraceIDMiddleware())

	// 4. CORS middleware (for development)
	e.Use(middleware.CORS())

	// 5. Recovery middleware (panic recovery)
	e.Use(middleware.Recover())

	// Health check endpoint (root level)
	e.GET(constants.HealthCheck, controllers.HealthCheck)

	// API group with version
	group := e.Group("/" + constants.MicroserviceName)
	v1 := group.Group(constants.V1)

	// Routing endpoints
	v1.POST(constants.Route, routingController.RouteTransaction)
	v1.GET(constants.Processors, routingController.GetProcessorHealth)
	v1.GET(constants.ProcessorByName, routingController.GetProcessorByName)
	v1.GET(constants.RoutingStats, routingController.GetRoutingStats)

	// Data management endpoints
	v1.POST(constants.TransactionsLoad, dataController.LoadTestData)
}
