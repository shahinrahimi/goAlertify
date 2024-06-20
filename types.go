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

type Alert struct {
	Id          string    `json:"id"`
	UserId      int64     `json:"user_id"`
	Symbol      string    `json:"symbol"`
	Description string    `json:"description"`
	TargetPrice float64   `json:"target_price"`
	StartPrice  float64   `json:"start_price"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

type Ticker struct {
	Symbol    string    `json:"name"`
	LivePrice float64   `json:"live_price"`
	DailyHigh float64   `json:"daily_high"`
	DailyLow  float64   `json:"daily_low"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Ticker) Update(livePrice, dailyHigh, dailyLow float64) error {
	t.UpdatedAt = time.Now().UTC()
	t.LivePrice = livePrice
	t.DailyHigh = dailyHigh
	t.DailyLow = dailyLow
	return nil
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

func NewAlert(userId int64, symbol, description string, targetPrice, startPrice float64) *Alert {
	return &Alert{
		Id:          fmt.Sprint("AL" + strconv.Itoa(rand.Int())),
		UserId:      userId,
		Description: description,
		Symbol:      symbol,
		TargetPrice: targetPrice,
		StartPrice:  startPrice,
		Active:      true,
		CreatedAt:   time.Now().UTC(),
	}
}

func NewTicker(symbol string, livePrice float64, dailyHigh float64, dailyLow float64) *Ticker {
	return &Ticker{
		Symbol:    symbol,
		LivePrice: livePrice,
		DailyHigh: dailyHigh,
		DailyLow:  dailyLow,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}
