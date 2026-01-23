package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"os"
	"strings"
	"github.com/lib/pq"
)

const RefreshTokenLength = 256;

type Language struct {
	LanguageId uint64
	Name string
}

type Hobby struct {
	HobbyId uint64
	Title string
}

type User struct {
	UserId uint64 `gorm:"primaryKey"`
	Username string
	PasswordHash []byte `gorm:"type:bytea"`
	RefreshToken string
	Description string // Bio
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

func (dbc *DBController) UpdateRefreshToken(user *User) error {
	return dbc.db.Raw(
		"UPDATE users SET refresh_token = ? WHERE refresh_token = ? RETURNING *",
		GenerateRefreshToken(),
		user.RefreshToken,
	).First(user).Error
}

type UserFilter struct {
	NameContains string `json:"name_contains"`
	HobbieIds []string `json:"hobbies"`
	KnownLanguageIds []string `json:"known_languages"`
	LearnLanguageIds []string `json:"learn_languages"`
}

type Users []struct {
	UserId uint64 `json:"user_id"`
	Username string `json:"username"`
	Description string `json:"description"`
	HobbyTitles pq.StringArray `json:"hobby_titles" gorm:"type:text[]"`
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
	`, filter.NameContains)

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
	
	err := dbc.db.Raw(query).Find(&result).Error
	return result, err
}

