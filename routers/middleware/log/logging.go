package log

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

// LoggingMiddleware provides simple request logging
func LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Log request details
			log.Printf("[%s] %s %s - Status: %d - Duration: %v",
				c.Request().Method,
				c.Request().URL.Path,
				c.Request().RemoteAddr,
				c.Response().Status,
				time.Since(start),
			)

			return err
		}
	}
}
