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
	UserId 					 uint64   `json:"user_id"`
	UsernameContains string   `json:"username_contains"`
	HobbieIds        []string `json:"hobbies"`
	KnownLanguageIds []string `json:"known_languages"`
	LearnLanguageIds []string `json:"learn_languages"`
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
		SELECT 
			u.user_id,
			u.username,
			u.description,
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
		WHERE u.username ILIKE '%%%v%%'
	`, filter.UsernameContains)

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
	return dbc.db.Exec(query, args...).Error
}

func (dbc *DBController) DeleteLanguage(userId, languageId uint64) error {
	return dbc.db.Exec("DELETE FROM user_languages WHERE user_id = ? AND language_id = ?", userId, languageId).Error
}

type Language struct {
	LanguageId uint64 `json:"language_id"`
	UserId     uint64
	IsKnown    bool `json:"is_known"`
}

func (dbc *DBController) AddLanguage(language *Language) error {
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
	HobbyId uint64 `json:"hobby_id"`
	UserId  uint64
}

func (dbc *DBController) AddHobby(hobby *Hobby) error {
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
	CreatedAt time.Time
}

type MessagesRequest struct {
	FromMessageId uint64
	Limit         uint64
}

func (dbc *DBController) GetMessagesInChat(request *MessagesRequest) ([]Message, error) {
	var messages []Message
	err := dbc.db.Raw(`
		WITH target_message AS (
			SELECT created_at, id 
			FROM messages 
			WHERE id = ?
		)
		SELECT m.*
		FROM messages m, target_message t
		WHERE 
			(m.created_at < t.created_at)
			OR (m.created_at = t.created_at AND m.id < t.id)
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT ?;
	`, request.FromMessageId, request.Limit).Find(&messages).Error
	return messages, err
}

func (dbc *DBController) CreateMessage(message *Message) error {
	return dbc.db.Exec(`
		INSERT INTO messages (message_id, sender_id, content, created_at) VALUES (?, ?, ?, ?)
	`, message.MessageId, message.SenderId, message.Content, message.CreatedAt).Error
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
