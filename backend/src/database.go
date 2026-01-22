package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"os"
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
//	Hobbies []Hobby
//	KnownLanguages []Language
//	LearnLanguages []Language
}

type DBController struct {
	db *gorm.DB
}

func (dbc *DBController) ConnectDB() error {
	dsn := fmt.Sprintf(
		"host=%v user=languasia password=1234 dbname=languasia port=5432 sslmode=disable TimeZone=Asia/Shanghai",
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

