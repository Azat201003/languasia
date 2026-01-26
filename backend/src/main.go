package main

import (
	"log/slog"
	"net/http"
	
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	// database
	dbc = new(DBController)
	
	if err := dbc.ConnectDB(); err != nil {
		panic(err)
	}

	// web socket
	wsh = NewHub()
	go wsh.Run()

	// server
  e := echo.New()

  e.Use(middleware.RequestLogger())
  e.Use(middleware.Recover())

  e.GET("/ping", func (c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.POST("/register", register)
	e.POST("/login", login)
	e.POST("/refresh", refresh)
	e.POST("/users", recieveFilteredUsers)
	e.GET("/ws", connectWebSocket)
  
	if err := e.Start(":8080"); err != nil {
    slog.Error("failed to start server", "error", err)
  }
}
