package controllers

import (
	"net/http"
	"voltarides/smart-router/models"
	"voltarides/smart-router/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// RoutingController handles routing-related requests
type RoutingController struct {
	service   *services.RoutingService
	validator *validator.Validate
}

// NewRoutingController creates a new routing controller
func NewRoutingController(service *services.RoutingService) *RoutingController {
	return &RoutingController{
		service:   service,
		validator: validator.New(),
	}
}

// RouteTransaction handles routing decision requests
func (rc *RoutingController) RouteTransaction(c echo.Context) error {
	var req models.RoutingRequest

	// Bind and validate request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
	}

	if err := rc.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "Request validation failed: " + err.Error(),
		})
	}

	// Check if simulation mode is enabled
	simulate := c.QueryParam("simulate") == "true"

	// Check if failover ranking is requested
	failover := c.QueryParam("failover") == "true"

	// Get routing decision
	var response *models.RoutingResponse
	var err error
	if failover {
		response, err = rc.service.SelectBestProcessorWithFailover(req, simulate)
	} else {
		response, err = rc.service.SelectBestProcessor(req, simulate)
	}
	if err != nil {
		// Check if it's an unsupported country error
		if err.Error() == "country "+req.Country+" not supported" {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "unsupported_country",
				Message: err.Error(),
			})
		}

		// Check if it's a no processors available error
		return c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "no_processors_available",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// GetProcessorHealth returns health stats for all processors
func (rc *RoutingController) GetProcessorHealth(c echo.Context) error {
	stats := rc.service.GetAllProcessorStats()

	return c.JSON(http.StatusOK, models.ProcessorHealthResponse{
		Processors: stats,
	})
}

// GetProcessorByName returns health stats for a specific processor
func (rc *RoutingController) GetProcessorByName(c echo.Context) error {
	name := c.Param("name")

	stat, err := rc.service.GetProcessorStats(name)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "processor_not_found",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, stat)
}

// GetRoutingStats returns routing decision statistics
func (rc *RoutingController) GetRoutingStats(c echo.Context) error {
	stats := rc.service.GetRoutingStats()

	return c.JSON(http.StatusOK, stats)
}
