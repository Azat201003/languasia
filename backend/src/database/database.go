package database

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"strings"
	"time"
)

var DBC *DBController

const RefreshTokenLength = 256

type User struct {
	UserId       uint64 `gorm:"primaryKey"`
	Username     string
	PasswordHash []byte `gorm:"type:bytea"`
	RefreshToken string
	Description  string // Bio
}

type DBController struct {
	db *gorm.DB
}

func (dbc *DBController) ConnectDB() error {
	dsn := fmt.Sprintf(
		"host=%v user=languasia password=1234 dbname=languasia port=5432 sslmode=disable",
		os.Getenv("DB_HOST"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	dbc.db = db
	return err
}

func (dbc *DBController) RegisterUser(user *User) error {
	return dbc.db.Exec(
		"INSERT INTO users (username, password_hash, refresh_token) VALUES (?, ?, ?)",
		user.Username,
		user.PasswordHash,
		GenerateRefreshToken(),
	).Error
}

func GenerateRefreshToken() string {
	bytes := make([]byte, RefreshTokenLength)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func (dbc *DBController) LoginUser(user *User) error {
	return dbc.db.Raw("SELECT * FROM users WHERE username = ? LIMIT 1",
		user.Username,
	).First(user).Error
}

func (dbc *DBController) UserByRefreshToken(user *User) error {
	return dbc.db.Raw(
		"SELECT * FRM users where refresh_token = ?",
		user.RefreshToken,
	).First(user).Error
}

type UserFilter struct {
	UserId           uint64   `json:"user_id"`
	SearchString     string   `json:"search_string"`
	HobbieIds        []string `json:"hobbies"`
	KnownLanguageIds []string `json:"known_languages"`
	LearnLanguageIds []string `json:"learn_languages"`
	PageSize         uint64   `json:"page_size"`
	PageNumber       uint64   `json:"page_number"`
}

type Users []struct {
	UserId             uint64         `json:"user_id"`
	Username           string         `json:"username"`
	Description        string         `json:"description"`
	HobbyTitles        pq.StringArray `json:"hobby_titles" gorm:"type:text[]"`
	KnownLanguageNames pq.StringArray `json:"known_language_names" gorm:"type:text[]"`
	LearnLanguageNames pq.StringArray `json:"learn_language_names" gorm:"type:text[]"`
}

func (dbc *DBController) RecieveFilteredUsers(filter *UserFilter) (Users, error) {
	var result Users

	query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT 
				u.user_id,
				u.username,
				u.description,
				similarity(u.username, '%v') AS sml1,
				similarity(u.description, '%v') AS sml2,
				COALESCE(h.hobby_titles, '{}') AS hobby_titles,
				COALESCE(kl.known_language_names, '{}') AS known_language_names,
				COALESCE(ll.learn_language_names, '{}') AS learn_language_names
			FROM users u
			LEFT JOIN LATERAL (
					SELECT array_agg(h.title) AS hobby_titles
					FROM user_hobbies uh
					JOIN hobbies h ON h.hobby_id = uh.hobby_id
					WHERE uh.user_id = u.user_id
			) h ON true
			LEFT JOIN LATERAL (
					SELECT array_agg(l.name) AS known_language_names
					FROM user_languages ul
					JOIN languages l ON l.language_id = ul.language_id
					WHERE ul.user_id = u.user_id AND ul.is_known = true
			) kl ON true
			LEFT JOIN LATERAL (
					SELECT array_agg(l.name) AS learn_language_names
					FROM user_languages ul
					JOIN languages l ON l.language_id = ul.language_id
					WHERE ul.user_id = u.user_id AND ul.is_known = false
			) ll ON true
			WHERE true
	`, filter.SearchString, filter.SearchString)

	if len(filter.HobbieIds) > 0 {
		query += fmt.Sprintf(`
			AND EXISTS (
					SELECT 1
					FROM user_hobbies ul2
					JOIN hobbies l2 ON l2.hobby_id = ul2.hobby_id
					WHERE ul2.user_id = u.user_id 
							AND l2.hobby_id IN (%v)
			)
		`, strings.Join(filter.HobbieIds, ", "))
	}
	if len(filter.KnownLanguageIds) > 0 {
		query += fmt.Sprintf(`
				AND EXISTS (
					SELECT 1
					FROM user_languages ul2
					JOIN languages l2 ON l2.language_id = ul2.language_id
					WHERE ul2.user_id = u.user_id 
						AND ul2.is_known = true 
						AND l2.language_id IN (%v)
				)
		`, strings.Join(filter.KnownLanguageIds, ", "))
	}
	if len(filter.KnownLanguageIds) > 0 {
		query += fmt.Sprintf(`
				AND EXISTS (
					SELECT 1
					FROM user_languages ul2
					JOIN languages l2 ON l2.language_id = ul2.language_id
					WHERE ul2.user_id = u.user_id 
						AND ul2.is_known = false
						AND l2.language_id IN (%v)
				)
		`, strings.Join(filter.LearnLanguageIds, ", "))
	}

	if filter.UserId != 0 {
		query += fmt.Sprintf(`
				AND u.user_id = %v
		`, filter.UserId)
	}

	query += fmt.Sprintf(`
		)
		ORDER BY 2*sml1+COALESCE(sml2,0) DESC, user_id ASC
		OFFSET %v*%v
		LIMIT %v
	`, max(filter.PageNumber, uint64(1))-1, filter.PageSize, max(filter.PageSize, uint64(1)))

	fmt.Println(query)

	err := dbc.db.Raw(query).Find(&result).Error
	return result, err
}

func (dbc *DBController) UpdadateUser(user *User) error {
	query := "UPDATE users "
	var args []any

	if user.Description != "" || len(user.PasswordHash) != 0 {
		query += "SET "
	}

	if len(user.PasswordHash) != 0 {
		query += "password_hash = ? "
		args = append(args, user.PasswordHash)
		if user.Description != "" {
			query += ", "
		}
	}

	if user.Description != "" {
		query += "description = ? "
		args = append(args, user.Description)
	}

	query += "WHERE user_id = ?"
	args = append(args, user.UserId)
    fmt.Println(query, args)
	return dbc.db.Exec(query, args...).Error
}

func (dbc *DBController) DeleteLanguage(userId, languageId uint64) error {
	return dbc.db.Exec("DELETE FROM user_languages WHERE user_id = ? AND language_id = ?", userId, languageId).Error
}

type Language struct {
	LanguageId uint64 `json:"language_id"`
	Name string `json:"name"`
}

func (dbc *DBController) GetLanguagesList() ([]Language, error) {
	var languages []Language
	err := dbc.db.Raw("SELECT * FROM languages").Find(&languages).Error
	return languages, err
}

type UserLanguage struct {
	LanguageId uint64 `json:"language_id"`
	UserId     uint64
	IsKnown    bool `json:"is_known"`
}

func (dbc *DBController) AddLanguage(language *UserLanguage) error {
	return dbc.db.Exec(
		"INSERT INTO user_languages (user_id, language_id, is_known) VALUES (?, ?, ?)",
		language.UserId,
		language.LanguageId,
		language.IsKnown,
	).Error
}

func (dbc *DBController) DeleteHobby(userId, hobbyId uint64) error {
	return dbc.db.Exec("DELETE FROM user_hobbies WHERE user_id = ? AND hobby_id = ?", userId, hobbyId).Error
}

type Hobby struct {
	Title string `json:"title"`
	HobbyId uint64 `json:"hobby_id"`
}

func (dbc *DBController) GetHobbiesList() ([]Hobby, error) {
	var hobbies []Hobby
	err := dbc.db.Raw("SELECT * FROM hobbies").Find(&hobbies).Error
	return hobbies, err
}

type UserHobby struct {
	HobbyId uint64 `json:"hobby_id"`
	UserId  uint64
}

func (dbc *DBController) AddHobby(hobby *UserHobby) error {
	return dbc.db.Exec(
		"INSERT INTO user_languages (user_id, hobby_id) VALUES (?, ?, ?)",
		hobby.UserId,
		hobby.HobbyId,
	).Error
}

func (dbc *DBController) DeleteUser(userId uint64) error {
	res := dbc.db.Exec("DELETE FROM users WHERE user_id = ?", userId)

	if res.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}

	return res.Error
}

type Message struct {
	MessageId uint64
	SenderId  uint64
	Content   string
	ChatId    uint64
	CreatedAt time.Time
}

type MessagesRequest struct {
	FromMessageId uint64 `json:"from_message_id"`
	Limit         uint64 `json:"limit"`
	ChatId        uint64 `json:"chat_id"`
}

func (dbc *DBController) GetMessagesInChat(request *MessagesRequest) ([]Message, error) {
	var messages []Message
	var err error
	if request.FromMessageId == 0 {
		err = dbc.db.Raw(`
			SELECT * FROM messages WHERE chat_id = ? ORDER BY created_at DESC, message_id DESC LIMIT ?
		`, request.ChatId, request.Limit).Find(&messages).Error
	} else {
		err = dbc.db.Raw(`
			WITH target_message AS (
				SELECT created_at, message_id
				FROM messages 
				WHERE message_id = ?
			)
			SELECT m.*
			FROM messages m, target_message t
			WHERE 
				((m.created_at < t.created_at)
				OR (m.created_at = t.created_at AND m.message_id < t.message_id))
				AND (m.chat_id = ?)
			ORDER BY m.created_at DESC, m.message_id DESC
			LIMIT ?;
		`, request.FromMessageId, request.ChatId, request.Limit).Find(&messages).Error
	}
	return messages, err
}

func (dbc *DBController) CreateMessage(message *Message) (time.Time, uint64, error) {
	var result Message
	err := dbc.db.Raw(`
		INSERT INTO messages (chat_id, sender_id, content) VALUES (?, ?, ?) RETURNING created_at, message_id
	`, message.ChatId, message.SenderId, message.Content).Scan(&result).Error
	return result.CreatedAt, result.MessageId, err
}

func (dbc *DBController) DeleteMessage(messageId uint64) error {
	return dbc.db.Delete(`
		DELETE FROM users WHERE message_id = ?
	`, messageId).Error
}

type Chat struct {
	ChatId uint64
	Title  string
	Type   string
}

func (dbc *DBController) CreateChat(chat *Chat) error {
	return dbc.db.Exec(`
		INSERT INTO chats (chat_id, title) VALUES (?, ?)
	`, chat.ChatId, chat.Title).Error
}

func (dbc *DBController) JoinChat(chatId, userId uint64) error {
	return dbc.db.Exec(`
		INSERT INTO chat_members (chat_id, user_id) VALUES (?, ?)
	`, chatId, userId).Error
}

func (dbc *DBController) MyChats(userId uint64) ([]Chat, error) {
	var chats []Chat
	err := dbc.db.Raw(`
		SELECT chats.chat_id, chats.title, chats.type FROM chat_members JOIN chats ON chat_members.chat_id = chats.chat_id WHERE chat_members.user_id = ?;
	`, userId).Find(&chats).Error
	return chats, err
}

func (dbc *DBController) GetChatMembers(chatId uint64) ([]uint64, error) {
	var members []uint64
	err := dbc.db.Raw(`
		SELECT user_id FROM chat_members WHERE chat_id = ?
	`, chatId).Find(&members).Error
	return members, err
}

func (dbc *DBController) LeaveChat(chatId, userId uint64) error {
	return dbc.db.Exec(`
		DELETE FROM chat_members WHERE user_id = ? AND chat_id = ?
	`, userId, chatId).Error
}

func (dbc *DBController) DeleteChat(chatId uint64) error {
	return dbc.db.Exec(`
		DELETE FROM chats WHERE chat_id = ?
	`, chatId).Error
}
