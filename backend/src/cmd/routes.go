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

	"github.com/Azat201003/languasia/backend/src/database"
)

type creditionals struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func register(c *echo.Context) error {
	request := new(creditionals)
	if err := c.Bind(request); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Bad data: %v", err))
		return err
	}

	if len(request.Password) < 2 {
		return c.String(http.StatusBadRequest, "Bad password: too short")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, fmt.Sprintf("Cannot get hash of password: %v", err.Error()))
		return err
	}

	err = dbc.RegisterUser(&database.User{
		Username:     request.Username,
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

	user := &database.User{
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
		"jwt_token":     jwt_token_string,
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
	user := &database.User{
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
		"jwt_token":     jwt_token_string,
	})
}

func recieveFilteredUsers(c *echo.Context) error {
	filter := new(database.UserFilter)

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

func deleteUser(c *echo.Context) error {
	userIdString := c.ParamOr("user_id", "0")

	userId, err := strconv.ParseUint(userIdString, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Cannot parse user_id: %v", err.Error()))
		return err
	}

	if c.Get("user_id") != userId {
		return c.String(http.StatusUnauthorized, "token user_id and put user_id not match")
	}

	err = dbc.DeleteUser(userId)
	if err != nil {
		c.String(http.StatusTeapot, fmt.Sprintf("Cannot delete user: %v", err.Error()))
		return err
	}

	return c.String(http.StatusOK, "User succefully deleted")
}

type updateUserRequest struct {
	Description string `json:"description"`
	Password    string `json:"password"`
}

func updateUser(c *echo.Context) error {
	userIdString := c.ParamOr("user_id", "0")

	userId, err := strconv.ParseUint(userIdString, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Cannot parse user_id: %v", err.Error()))
		return err
	}

	if c.Get("user_id") != userId {
		return c.String(http.StatusUnauthorized, fmt.Sprintf("token user_id and put user_id not match: %v != %v", userId, c.Get("user_id")))
	}

	request := new(updateUserRequest)
	if err := c.Bind(request); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Cannot parse user data: %v", err.Error()))
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, fmt.Sprintf("Cannot get hash of password: %v", err.Error()))
		return err
	}

	err = dbc.UpdadateUser(&database.User{
		UserId:       userId,
		PasswordHash: passwordHash,
		Description:  request.Description,
	})
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "Succesfully updated")
}

func addLanguage(c *echo.Context) error {
	request := new(database.Language)
	if err := c.Bind(request); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Cannot parse language data: %v", err.Error()))
		return err
	}

	var err error
	request.UserId, err = strconv.ParseUint(c.ParamOr("user_id", "0"), 10, 64)
	if err != nil {
		return err
	}

	if c.Get("user_id") != request.UserId {
		return c.String(http.StatusUnauthorized, "token user_id and put user_id not match")
	}

	return dbc.AddLanguage(request)
}

func deleteLanguage(c *echo.Context) error {
	userId, err := strconv.ParseUint(c.ParamOr("user_id", "0"), 10, 64)
	if err != nil {
		return err
	}

	if c.Get("user_id") != userId {
		return c.String(http.StatusUnauthorized, "token user_id and put user_id not match")
	}

	languageId, err := strconv.ParseUint(c.ParamOr("language_id", "0"), 10, 64)
	if err != nil {
		return err
	}

	return dbc.DeleteLanguage(userId, languageId)
}

func addHobby(c *echo.Context) error {
	request := new(database.Hobby)
	if err := c.Bind(request); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Cannot parse language data: %v", err.Error()))
		return err
	}

	var err error
	request.UserId, err = strconv.ParseUint(c.ParamOr("user_id", "0"), 10, 64)
	if err != nil {
		return err
	}

	if c.Get("user_id") != request.UserId {
		return c.String(http.StatusUnauthorized, "token user_id and put user_id not match")
	}

	return dbc.AddHobby(request)
}

func deleteHobby(c *echo.Context) error {
	userId, err := strconv.ParseUint(c.ParamOr("user_id", "0"), 10, 64)
	if err != nil {
		return err
	}

	if c.Get("user_id") != userId {
		return c.String(http.StatusUnauthorized, "token user_id and put user_id not match")
	}

	hobbyId, err := strconv.ParseUint(c.ParamOr("hobby_id", "0"), 10, 64)
	if err != nil {
		return err
	}

	return dbc.DeleteHobby(userId, hobbyId)
}
