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
}

func GetCreateUsersTable() string {
	return `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL UNIQUE,
		username TEXT NOT NULL,
		firstname TEXT NOT NULL,
		lastname TEXT NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
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
	}, nil
}
