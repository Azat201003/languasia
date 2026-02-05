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

const MAX_PAGE_SIZE = 1000

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

	err = database.DBC.RegisterUser(&database.User{
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
	err := database.DBC.LoginUser(user)
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
	err := database.DBC.UserByRefreshToken(user)
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

	if filter.PageSize > MAX_PAGE_SIZE {
		return c.String(http.StatusBadRequest, "Page size limit exceeded")
	}

	users, err := database.DBC.RecieveFilteredUsers(filter)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, users)
}

func connectWebSocket(c *echo.Context) (err error) {
	//defer fmt.Println(err.Error())

	if c.Get("user_id") == nil {
		return c.String(http.StatusUnauthorized, "Token user_id is nil")
	}

	if _, ok := c.Get("user_id").(uint64); !ok {
		return c.String(http.StatusUnauthorized, "Cannot parse user_id")
	}

	fmt.Println("id: ", c.Get("user_id").(uint64))

	users, err := database.DBC.RecieveFilteredUsers(&database.UserFilter{UserId: c.Get("user_id").(uint64)})
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Cannot find user: %v", err.Error()))
	}

	fmt.Println(users[0])

	if len(users) == 0 {
		return c.String(http.StatusBadRequest, "Cannot find user")
	}

	user := users[0]

	return wsh.ConnectWebSocket(c.Response(), c.Request(), database.User{
		UserId:      user.UserId,
		Username:    user.Username,
		Description: user.Description,
	})
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

	err = database.DBC.DeleteUser(userId)
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

	if len(request.Password) < 2 {
		return c.String(http.StatusBadRequest, "Bad password: too short")
	}

    fmt.Printf("Update user password: %v\n", request.Password)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, fmt.Sprintf("Cannot get hash of password: %v", err.Error()))
		return err
	}

	err = database.DBC.UpdadateUser(&database.User{
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

	return database.DBC.AddLanguage(request)
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

	return database.DBC.DeleteLanguage(userId, languageId)
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

	return database.DBC.AddHobby(request)
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

	return database.DBC.DeleteHobby(userId, hobbyId)
}
