package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type User struct {
	Id        string    `json:"id"`
	UserId    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Firstname string    `json:"fistname"`
	Lastname  string    `json:"lastname"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	IsAdmin   bool      `json:"is_admin"`
}

func GetCreateUsersTable() string {
	return `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL UNIQUE,
		username TEXT NOT NULL,
		firstname TEXT NOT NULL,
		lastname TEXT NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		is_admin BOOLEAN
	);`
}

func NewUser(user_id int64, username, firstname, lastname, password string) (*User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	return &User{
		Id:        fmt.Sprint("GU" + strconv.Itoa(rand.Int())),
		UserId:    user_id,
		Username:  username,
		Firstname: firstname,
		Lastname:  lastname,
		Password:  hashedPassword,
		CreatedAt: time.Now().UTC(),
		IsAdmin:   false,
	}, nil
}

func NewAdmin(user_id int64, password string) (*User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	return &User{
		Id:        fmt.Sprint("AD99999999999"),
		UserId:    user_id,
		Username:  "admin",
		Firstname: "admin",
		Lastname:  "admin",
		Password:  hashedPassword,
		CreatedAt: time.Now().UTC(),
		IsAdmin:   false,
	}, nil
}

func (u *User) toTelegramString() string {
	return fmt.Sprintf("User ID: %d\nChat ID: %d\nUsername: %s\nFistname: %s\nLastname: %s\nPassword: %s\nCreated At: %s",
		u.UserId, u.UserId, u.Username, u.Firstname, u.Lastname, u.Password, u.CreatedAt.Format(time.RFC3339))
}
