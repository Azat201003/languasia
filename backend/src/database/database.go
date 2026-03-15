package database

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBC *DBController

const RefreshTokenLength = 256

type User struct {
	UserId       uint64 `gorm:"primaryKey"`
	Username     string
	Nickname     string
	Color        string
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
		"INSERT INTO users (username, nickname, password_hash, refresh_token) VALUES (?, ?, ?, ?)",
		user.Username,
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
		"SELECT * FROM users where refresh_token = ?",
		user.RefreshToken,
	).First(user).Error
}

type UserFilter struct {
	UserId           uint64   `json:"user_id"`
	SearchString     string   `json:"search_string"`
	HobbieIds        []uint64 `json:"hobbies"`
	KnownLanguageIds []uint64 `json:"known_languages"`
	LearnLanguageIds []uint64 `json:"learn_languages"`
	PageSize         uint64   `json:"page_size"`
	PageNumber       uint64   `json:"page_number"`
}

type Users []struct {
	UserId           uint64        `json:"user_id"`
	Username         string        `json:"username"`
	Nickname         string        `json:"nickname"`
	Color            string        `json:"color"`
	Description      string        `json:"description"`
	HobbyIds         pq.Int64Array `json:"hobby_title_ids" gorm:"type:serial[]"`
	KnownLanguageIds pq.Int64Array `json:"known_language_ids" gorm:"type:serial[]"`
	LearnLanguageIds pq.Int64Array `json:"learn_language_ids" gorm:"type:serial[]"`
}

func (dbc *DBController) RecieveFilteredUsers(filter *UserFilter) (Users, error) {
	var result Users

	// Build array strings for WHERE conditions
	var hobbyArrayStr, knownLangArrayStr, learnLangArrayStr string

	if len(filter.HobbieIds) > 0 {
		hobbyIdStrings := make([]string, len(filter.HobbieIds))
		for i := 0; i < len(filter.HobbieIds); i++ {
			hobbyIdStrings[i] = strconv.FormatUint(filter.HobbieIds[i], 10)
		}
		hobbyArrayStr = "ARRAY[" + strings.Join(hobbyIdStrings, ",") + "]"
	} else {
		hobbyArrayStr = "ARRAY[]::bigint[]"
	}

	if len(filter.KnownLanguageIds) > 0 {
		knownLanguageIdStrings := make([]string, len(filter.KnownLanguageIds))
		for i := 0; i < len(filter.KnownLanguageIds); i++ {
			knownLanguageIdStrings[i] = strconv.FormatUint(filter.KnownLanguageIds[i], 10)
		}
		knownLangArrayStr = "ARRAY[" + strings.Join(knownLanguageIdStrings, ",") + "]"
	} else {
		knownLangArrayStr = "ARRAY[]::bigint[]"
	}

	if len(filter.LearnLanguageIds) > 0 {
		learnLanguageIdStrings := make([]string, len(filter.LearnLanguageIds))
		for i := 0; i < len(filter.LearnLanguageIds); i++ {
			learnLanguageIdStrings[i] = strconv.FormatUint(filter.LearnLanguageIds[i], 10)
		}
		learnLangArrayStr = "ARRAY[" + strings.Join(learnLanguageIdStrings, ",") + "]"
	} else {
		learnLangArrayStr = "ARRAY[]::bigint[]"
	}

	query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT
				u.user_id,
				u.username,
				u.description,
				u.nickname,
				u.color,
				similarity(u.username, '%v') AS sml1,
				similarity(u.nickname, '%v') AS sml2,
				similarity(u.description, '%v') AS sml3,
				COALESCE(h.hobby_ids, '{}') AS hobby_ids,
				COALESCE(kl.known_language_ids, '{}') AS known_language_ids,
				COALESCE(ll.learn_language_ids, '{}') AS learn_language_ids,
				-- Count matching hobbies
				(
					SELECT COUNT(*) 
					FROM unnest(COALESCE(h.hobby_ids, '{}')) AS user_hobby_id
					WHERE user_hobby_id = ANY(%s)
				) AS matched_hobbies_count,
				-- Count matching known languages
				(
					SELECT COUNT(*) 
					FROM unnest(COALESCE(kl.known_language_ids, '{}')) AS user_known_lang_id
					WHERE user_known_lang_id = ANY(%s)
				) AS matched_known_languages_count,
				-- Count matching learn languages
				(
					SELECT COUNT(*) 
					FROM unnest(COALESCE(ll.learn_language_ids, '{}')) AS user_learn_lang_id
					WHERE user_learn_lang_id = ANY(%s)
				) AS matched_learn_languages_count
			FROM users u
			LEFT JOIN LATERAL (
				SELECT array_agg(h.hobby_id) AS hobby_ids
				FROM user_hobbies uh
				JOIN hobbies h ON h.hobby_id = uh.hobby_id
				WHERE uh.user_id = u.user_id
			) h ON true
			LEFT JOIN LATERAL (
				SELECT array_agg(l.language_id) AS known_language_ids
				FROM user_languages ul
				JOIN languages l ON l.language_id = ul.language_id
				WHERE ul.user_id = u.user_id AND ul.is_known = true
			) kl ON true
			LEFT JOIN LATERAL (
				SELECT array_agg(l.language_id) AS learn_language_ids
				FROM user_languages ul
				JOIN languages l ON l.language_id = ul.language_id
				WHERE ul.user_id = u.user_id AND ul.is_known = false
			) ll ON true
			WHERE true
	`,
		filter.SearchString, filter.SearchString, filter.SearchString,
		hobbyArrayStr, knownLangArrayStr, learnLangArrayStr)

	// Add WHERE conditions
	if len(filter.HobbieIds) > 0 {
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM user_hobbies uh2
				JOIN hobbies h2 ON h2.hobby_id = uh2.hobby_id
				WHERE uh2.user_id = u.user_id
				AND h2.hobby_id = ANY(%s)
			)
		`, hobbyArrayStr)
	}
	if len(filter.KnownLanguageIds) > 0 {
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM user_languages ul2
				JOIN languages l2 ON l2.language_id = ul2.language_id
				WHERE ul2.user_id = u.user_id
				AND ul2.is_known = true
				AND l2.language_id = ANY(%s)
			)
		`, knownLangArrayStr)
	}
	if len(filter.LearnLanguageIds) > 0 {
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM user_languages ul2
				JOIN languages l2 ON l2.language_id = ul2.language_id
				WHERE ul2.user_id = u.user_id
				AND ul2.is_known = false
				AND l2.language_id = ANY(%s)
			)
		`, learnLangArrayStr)
	}

	if filter.UserId != 0 {
		query += fmt.Sprintf(`
			AND u.user_id = %v
		`, filter.UserId)
	}

	query += `
		)
	`

	if len(filter.SearchString) > 0 {
		query += `
			WHERE (
				COALESCE(sml1,0) > 0.2 OR
				COALESCE(sml2,0) > 0.2 OR
				COALESCE(sml3,0) > 0.07
			)
		`
	}

	// Add ORDER BY with sum of matching counts
	query += fmt.Sprintf(`
		ORDER BY 
			2*sml1 + COALESCE(sml2,0) + COALESCE(sml3,0) DESC, 
			(matched_hobbies_count + matched_known_languages_count + matched_learn_languages_count) DESC,
			user_id ASC
		OFFSET %v*%v
		LIMIT %v
	`, max(filter.PageNumber, uint64(1))-1, filter.PageSize, max(filter.PageSize, uint64(1)))

	//	fmt.Println(query)

	err := dbc.db.Raw(query).Find(&result).Error
	return result, err
}

func (dbc *DBController) UpdadateUser(user *User, newLanguageIds []UserLanguage, oldLanguageIds []UserLanguage, newHobbyIds []UserHobby, oldHobbyIds []UserHobby) error {
	query := "UPDATE users "
	var args []any
	var updateStrings []string

	if len(user.PasswordHash) != 0 {
		updateStrings = append(updateStrings, "password_hash = ?")
		args = append(args, user.PasswordHash)
	}

	if user.Description != "" {
		updateStrings = append(updateStrings, "description = ?")
		args = append(args, user.Description)
	}

	if user.Nickname != "" {
		updateStrings = append(updateStrings, "nickname = ?")
		args = append(args, user.Nickname)
	}

	if user.Color != "" {
		updateStrings = append(updateStrings, "color = ?")
		args = append(args, user.Color)
	}

	var err error

	if len(updateStrings) > 0 {
		query += "SET "

		query += strings.Join(updateStrings, ", ")

		query += " WHERE user_id = ?"
		args = append(args, user.UserId)
		fmt.Println(query, args)

		err = dbc.db.Exec(query, args...).Error
	}

	for _, language := range newLanguageIds {
		language.UserId = user.UserId
		err = errors.Join(err, dbc.AddLanguage(&language))
	}

	for _, language := range oldLanguageIds {
		language.UserId = user.UserId
		err = errors.Join(err, dbc.DeleteLanguage(&language))
	}

	for _, hobby := range newHobbyIds {
		hobby.UserId = user.UserId
		err = errors.Join(err, dbc.AddHobby(&hobby))
	}

	for _, hobby := range oldHobbyIds {
		hobby.UserId = user.UserId
		err = errors.Join(err, dbc.DeleteHobby(&hobby))
	}

	return err
}

func (dbc *DBController) DeleteLanguage(language *UserLanguage) error {
	return dbc.db.Exec("DELETE FROM user_languages WHERE user_id = ? AND language_id = ?", language.UserId, language.LanguageId).Error
}

type Language struct {
	LanguageId uint64 `json:"language_id"`
	Name       string `json:"name"`
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

func (dbc *DBController) DeleteHobby(hobby *UserHobby) error {
	return dbc.db.Exec("DELETE FROM user_hobbies WHERE user_id = ? AND hobby_id = ?", hobby.UserId, hobby.HobbyId).Error
}

type Hobby struct {
	Title   string `json:"title"`
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
		"INSERT INTO user_hobbies (user_id, hobby_id) VALUES (?, ?)",
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
	MessageId uint64    `json:"message_id"`
	SenderId  uint64    `json:"sender_id"`
	Content   string    `json:"content"`
	ChatId    uint64    `json:"chat_id"`
	CreatedAt time.Time `json:"created_at"`
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
	ChatId      uint64        `json:"chat_id"`
	Title       string        `json:"title"`
	Type        string        `json:"type"`
	Color       string        `json:"color"`
	MemberIds   pq.Int64Array `json:"member_ids" gorm:"type:serial[]"`
	GoalId      uint64        `json:"goal_id"` // For type="Direct"
	LastMessage *Message      `json:"last_message" gorm:"-"`
}

func (dbc *DBController) CreateChat(chat *Chat) (chatId uint64, err error) {
	err = dbc.db.Raw(`
		INSERT INTO chats (title) VALUES (?) RETURNING chat_id
	`, chat.Title).Scan(&chatId).Error
	return
}

func (dbc *DBController) JoinChat(chatId, userId uint64) error {
	return dbc.db.Exec(`
		INSERT INTO chat_members (chat_id, user_id) VALUES (?, ?)
	`, chatId, userId).Error
}

func (dbc *DBController) MyChats(userId uint64) ([]Chat, error) {
	var chats []Chat
	err := dbc.db.Raw(`
		WITH user_chats AS (
			SELECT DISTINCT chat_id
			FROM chat_members
			WHERE user_id = ?
		)
		SELECT
			c.chat_id,
			c.title,
			c.type,
			(
				SELECT array_agg(user_id)
				FROM chat_members cm
				WHERE cm.chat_id = c.chat_id
			) as member_ids
		FROM user_chats uc
		JOIN chats c ON uc.chat_id = c.chat_id;
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

func (dbc *DBController) GetChat(chatId uint64) (*Chat, error) {
	var chat *Chat
	err := dbc.db.Raw(`
		SELECT
			c.chat_id,
			c.title,
			c.type,
			(
				SELECT array_agg(user_id)
				FROM chat_members cm
				WHERE cm.chat_id = c.chat_id
			) as member_ids
		FROM chats c
		WHERE ? = c.chat_id;
	`, chatId).Find(&chat).Error
	return chat, err
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
