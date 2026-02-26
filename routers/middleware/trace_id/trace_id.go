package trace_id

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TraceIDMiddleware adds a trace ID to each request
func TraceIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if trace ID already exists in headers
			traceID := c.Request().Header.Get("X-Trace-Id")
			if traceID == "" {
				// Generate new trace ID
				traceID = uuid.New().String()
			}

			// Set trace ID in response header
			c.Response().Header().Set("X-Trace-Id", traceID)

			// Store in context for potential use
			c.Set("trace_id", traceID)

			return next(c)
		}
	}
}
