package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/Azat201003/languasia/backend/src/database"
	"github.com/Azat201003/languasia/backend/src/websocket"
)

func authMiddlware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) (err error) {
		defer next(c)
		tokenString := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

		fmt.Println(c.Request().Header.Get("Authorization"), tokenString)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			publicKey, err := os.ReadFile("public_key.pem")
			if err != nil {
				return nil, err
			}
			return jwt.ParseEdPublicKeyFromPEM(publicKey)
		}, jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}))

		if err != nil {
			return err
		}

		userIdString, err := token.Claims.GetIssuer()
		if err != nil {
			return err
		}

		userId, err := strconv.ParseUint(userIdString, 10, 64)

		if err != nil {
			return err
		}

		c.Set("user_id", userId)

		return nil
	}
}

var wsh *websocket.WebSocketHub
var dbc *database.DBController

func main() {
	// database
	dbc = new(database.DBController)

	if err := dbc.ConnectDB(); err != nil {
		panic(err)
	}

	// web socket
	wsh = websocket.NewHub()
	go wsh.Run()

	// server
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(authMiddlware)

	e.GET("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.POST("/register", register)
	e.POST("/login", login)
	e.POST("/refresh", refresh)
	e.GET("/users", recieveFilteredUsers)
	//e.GET("/users/:user_id", getUser)
	e.DELETE("/users/:user_id", deleteUser)
	e.PATCH("/users/:user_id", updateUser)
	e.POST("/users/:user_id/languages", addLanguage)
	e.DELETE("/users/:user_id/languages/:language_id", deleteLanguage)
	e.POST("/users/:user_id/hobbies", addHobby)
	e.DELETE("/users/:user_id/hobbies/:hobby_id", deleteHobby)

	e.GET("/ws", connectWebSocket)

	if err := e.Start(":8080"); err != nil {
		slog.Error("failed to start server", "error", err)
	}
}
