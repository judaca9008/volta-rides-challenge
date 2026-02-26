package controllers

import (
	"net/http"
	"voltarides/smart-router/models"

	"github.com/labstack/echo/v4"
)

// HealthCheck handles health check requests
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, models.HealthResponse{
		Status: "Ok",
	})
}
