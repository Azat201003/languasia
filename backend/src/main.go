package main

import (
  "github.com/labstack/echo/v5"
  "github.com/labstack/echo/v5/middleware"
  "log/slog"
	"net/http"
)

func main() {
  e := echo.New()

  e.Use(middleware.RequestLogger()) // use the RequestLogger middleware with slog logger
  e.Use(middleware.Recover()) // recover panics as errors for proper error handling

  e.GET("/", hello)

  // Start server
  if err := e.Start(":8080"); err != nil {
    slog.Error("failed to start server", "error", err)
  }
}

// Handler
func hello(c *echo.Context) error {
  return c.String(http.StatusOK, "Hello, World!")
}
