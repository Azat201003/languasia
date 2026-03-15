package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
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
        "user_id":       user.UserId,
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
	Nickname    string `json:"nickname"`
	Color       string `json:"color"`
	AddHobbyIds []database.UserHobby `json:"add_hobbies"`
	DeleteHobbyIds []database.UserHobby `json:"delete_hobbies"`
	AddLanguageIds []database.UserLanguage `json:"add_languages"`
	DeleteLanguageIds []database.UserLanguage `json:"delete_languages"`
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

	if len(request.Password) < 2 && request.Password != "" {
		return c.String(http.StatusBadRequest, "Bad password: too short")
	}

	fmt.Printf("Update user password: %v\n", request.Password)
	
	var passwordHash []byte

	if request.Password != "" {
		passwordHash, err = bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			c.String(http.StatusUnprocessableEntity, fmt.Sprintf("Cannot get hash of password: %v", err.Error()))
			return err
		}
	}

	err = database.DBC.UpdadateUser(&database.User{
		UserId:       userId,
		PasswordHash: passwordHash,
		Description:  request.Description,
        Nickname:     request.Nickname,
        Color:        request.Color,
	}, request.AddLanguageIds, request.DeleteLanguageIds, request.AddHobbyIds, request.DeleteHobbyIds)

	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "Succesfully updated")
}

func addLanguage(c *echo.Context) error {
	request := new(database.UserLanguage)
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

	return database.DBC.DeleteLanguage(&database.UserLanguage{UserId: userId, LanguageId: languageId})
}

func addHobby(c *echo.Context) error {
	request := new(database.UserHobby)
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

	return database.DBC.DeleteHobby(&database.UserHobby{UserId: userId, HobbyId: hobbyId})
}

func getListLanguages(c *echo.Context) error {
	languages, err := database.DBC.GetLanguagesList()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot get list of languages: %v", err.Error()))
	}
	return c.JSON(http.StatusOK, languages)
}

func getListHobbies(c *echo.Context) error {
	hobbies, err := database.DBC.GetHobbiesList()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Cannot get list of hobbies: %v", err.Error()))
	}
	return c.JSON(http.StatusOK, hobbies)
}

func getChatList(c *echo.Context) error {	
	userIdString := c.ParamOr("user_id", "0")

	userId, err := strconv.ParseUint(userIdString, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Cannot parse user_id: %v", err.Error()))
		return err
	}

	if c.Get("user_id") != userId {
		return c.String(http.StatusUnauthorized, fmt.Sprintf("Token's user_id and put user_id not match: %v != %v", userId, c.Get("user_id")))
	}

	chats, err := database.DBC.MyChats(userId)
	for i := range chats {
		if chats[i].Type == "Direct" {
			_ = completeDirectChat(userId, &chats[i])
			
			if len(chats[i].MemberIds) > 2 {
				chats[i].MemberIds = chats[i].MemberIds[0:2]
			}
		}
		messages, err := database.DBC.GetMessagesInChat(&database.MessagesRequest{
			Limit: 1,
			ChatId: chats[i].ChatId,
		})
		if len(messages) < 1 || err != nil {
			fmt.Println(err)
			continue
		} else {
			fmt.Println("messages: ", messages)
			chats[i].LastMessage = &messages[0]
		}
	}
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Cannot get chat list: %v", err.Error()))
	}
	return c.JSON(http.StatusOK, chats)
}

func completeDirectChat(myUserId uint64, chat *database.Chat) error {
	memberIds, err := database.DBC.GetChatMembers(chat.ChatId)
	if err != nil {
		return err
	}
	if len(memberIds) < 2 {
		return errors.New("...")
	}
	fmt.Println("Member ids: ", memberIds)
	var memberId uint64
	if memberIds[0] == myUserId {
		memberId = memberIds[1]
	} else {
		memberId = memberIds[0]
	}
	members, err := database.DBC.RecieveFilteredUsers(&database.UserFilter{UserId: memberId})
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return errors.New("...")
	}
	chat.Color = members[0].Color
	chat.Title = members[0].Nickname
	chat.GoalId = members[0].UserId
	fmt.Println("Member: ", members[0])
	return nil 
}

func getChat(c *echo.Context) error {	
	userId, ok := c.Get("user_id").(uint64)
	if !ok {
		return c.String(http.StatusBadRequest, "Cannot parse user_id from token")
	}

	chatIdString := c.ParamOr("chat_id", "0")

	chatId, err := strconv.ParseUint(chatIdString, 10, 64)
	
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Cannot parse chat_id: %v", err.Error()))
		return err
	}

	chat, err := database.DBC.GetChat(chatId)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Cannot get chat: %v", err.Error()))
	}

	if chat.Type == "Direct" {
		_ = completeDirectChat(userId, chat)
	}

	var memberIds []int64
	memberIds = []int64(chat.MemberIds)
	if !slices.Contains(memberIds, int64(userId)) {
		return c.String(http.StatusForbidden, "User is not member of this chat")
	}

	return c.JSON(http.StatusOK, chat)
}

func createChat(c *echo.Context) error {
	request := new(database.Chat)
	if err := c.Bind(request); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Cannot parse chat: %v", err.Error()))
		return err
	}

	userId, ok := c.Get("user_id").(uint64)
	if !ok {
		return c.String(http.StatusUnauthorized, "User id in autharization token not found")
	}

	if request.Type == "Direct" {
		chats, err := database.DBC.MyChats(userId)
		if err != nil {
			return c.String(http.StatusBadRequest, "Cannot get your chat")
		}
		for _, chat := range chats {
			if chat.Type == "Direct" && len(chat.MemberIds) >= 2 && (
				chat.MemberIds[0] == int64(userId) && chat.MemberIds[1] == int64(request.GoalId) ||
				chat.MemberIds[1] == int64(userId) && chat.MemberIds[0] == int64(request.GoalId)) {
				return c.String(http.StatusConflict, "Direct chat with these members already exists")
			}
		}
		if request.GoalId == userId {
			return c.String(http.StatusBadRequest, "Cannot create chat with yourself")
		}
	}

	chatId, err := database.DBC.CreateChat(request)
	if err != nil {
		return c.String(http.StatusBadRequest, "Cannot create chat")
	}

	err = database.DBC.JoinChat(chatId, userId)

	if err != nil {
		database.DBC.DeleteChat(chatId)
		return c.String(http.StatusTeapot, "I wanna make tea when see that I cannot add you to chat :(. (btw chat created and it is just trash)")
	}
	
	if request.Type == "Direct" {
		err = database.DBC.JoinChat(chatId, request.GoalId)
		if err != nil {
			database.DBC.DeleteChat(chatId)
			return c.String(http.StatusTeapot, "I wanna make tea when see that I cannot add user to chat :(. (btw chat created and it is just trash)")
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"chat_id": chatId,
	})
}

func addMember(c *echo.Context) error {	
	myUserId, ok := c.Get("user_id").(uint64)
	if myUserId == uint64(0) || myUserId == 0 || !ok {
		return c.String(http.StatusUnauthorized, "my user id = 0")
	}

	userId, err := strconv.ParseUint(c.ParamOr("user_id", "0"), 10, 64)
	if err != nil {
		return err
	}

	chatId, err := strconv.ParseUint(c.ParamOr("chat_id", "0"), 10, 64)
	if err != nil {
		return err
	}

	chat, err := database.DBC.GetChat(chatId)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Cannot get members: %v", err.Error()))
	}

	var memberIds []int64
	memberIds = []int64(chat.MemberIds)
	if !slices.Contains(memberIds, int64(myUserId)) {
		return c.String(http.StatusForbidden, "User is not member of this chat")
	}
	
	err = database.DBC.JoinChat(chatId, userId)

	if err != nil {
		return c.String(http.StatusTeapot, "I wanna make tea when see that I cannot add user to chat :(.")
	}
	

	return c.NoContent(http.StatusOK)
}

func getMyUser(c *echo.Context) error {
	userId, ok := c.Get("user_id").(uint64)
	if !ok {
		return c.String(http.StatusUnauthorized, "Cannot parse user_id")
	}
	users, err := database.DBC.RecieveFilteredUsers(&database.UserFilter{
		UserId: userId,
	})

	if err != nil {
		return c.String(http.StatusUnauthorized, fmt.Sprintf("Cannot find user: %v", err.Error()))
	}

	if len(users) < 1 {
		return c.String(http.StatusUnauthorized, "No users found")
	}
	user := users[0]
	return c.JSON(http.StatusOK, user)
}

