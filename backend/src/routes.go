package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

type creditionals struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func register(c *echo.Context) error {
	request := new(creditionals)
	if err := c.Bind(request); err != nil {
		return err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, fmt.Sprintf("Cannot get hash of password: %v", err.Error()))
		return err
	}

	err = dbc.RegisterUser(&User{
		Username: request.Username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		c.String(http.StatusUnprocessableEntity, fmt.Sprintf("Cannot create user: %v", err.Error()))
		return err
	}
	return c.String(http.StatusOK, "Registered") // TODO I can answer smth
}

func login(c *echo.Context) error {	
	request := new(creditionals)
	if err := c.Bind(request); err != nil {
		return err
	}

	user := &User{
		Username: request.Username,
	}
	fmt.Println("!!USER: ", user)
	err := dbc.LoginUser(user)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, fmt.Sprintf("Cannot find user: %v", err.Error()))
		return err
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(request.Password))
	if err != nil {
		c.String(http.StatusUnauthorized, fmt.Sprintf("Password not match: %v", err.Error()))
		return err
	}

	privateKey, err := os.ReadFile("private_key.pem")
	
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot read private_key.pem: %v", err.Error()))
	} 

	userIdString := strconv.FormatUint(user.UserId, 10)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    userIdString,
	}

	jwt_token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	parsedPrivateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKey)

	if err != nil {
		c.String(http.StatusUnauthorized, fmt.Sprintf("Cannot parse PEM from private key: %v", err.Error()))
		return err
	}
	jwt_token_string, err := jwt_token.SignedString(parsedPrivateKey)

	if err != nil {
		c.String(http.StatusUnauthorized, fmt.Sprintf("Cannot sign jwt token: %v", err.Error()))
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{
		"refresh_token": user.RefreshToken,
		"jwt_token": jwt_token_string,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func refresh(c *echo.Context) error {
	request := new(refreshRequest)
	if err := c.Bind(request); err != nil {
		return err
	}
	user := &User{
		RefreshToken: request.RefreshToken,
	}
	err := dbc.UserByRefreshToken(user)
	if err != nil {
		c.String(http.StatusUnauthorized, "Not valid refresh token")
		return err
	}
	
	privateKey, err := os.ReadFile("private_key.pem")
	
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot read private_key.pem: %v", err.Error()))
	} 

	userIdString := strconv.FormatUint(user.UserId, 10)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    userIdString,
	}

	jwt_token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	parsedPrivateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKey)

	if err != nil {
		c.String(http.StatusUnauthorized, fmt.Sprintf("Cannot parse PEM from private key: %v", err.Error()))
		return err
	}
	jwt_token_string, err := jwt_token.SignedString(parsedPrivateKey)

	if err != nil {
		c.String(http.StatusUnauthorized, fmt.Sprintf("Cannot sign jwt token: %v", err.Error()))
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{
		"refresh_token": user.RefreshToken,
		"jwt_token": jwt_token_string,
	})
}

func recieveFilteredUsers(c *echo.Context) error {
	filter := new(UserFilter)

	if err := c.Bind(filter); err != nil {
		return err
	}

	users, err := dbc.RecieveFilteredUsers(filter)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusFound, users)
}

func connectWebSocket(c *echo.Context) error {
	return wsh.ConnectWebSocket(c.Response(), c.Request())
}

