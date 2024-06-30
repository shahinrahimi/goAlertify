package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Alert struct {
	Id          string    `json:"id"`
	UserId      int64     `json:"user_id"`
	Number      int32     `json:"number"`
	Symbol      string    `json:"symbol"`
	Description string    `json:"description"`
	TargetPrice float64   `json:"target_price"`
	StartPrice  float64   `json:"start_price"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func GetCreateAlertsTable() string {
	return `CREATE TABLE IF NOT EXISTS alerts (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		number INTEGER,
		symbol TEXT NOT NULL,
		description TEXT,
		target_price REAL,
		start_price REAL,
		active BOOLEAN,
		updated_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (user_id)
	);`
}

func NewAlert(userId int64, symbol, description string, targetPrice, startPrice float64) *Alert {
	return &Alert{
		Id:          fmt.Sprint("AL" + strconv.Itoa(rand.Int())),
		UserId:      userId,
		Number:      9999,
		Description: description,
		Symbol:      symbol,
		TargetPrice: targetPrice,
		StartPrice:  startPrice,
		Active:      true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}

func (a *Alert) ToString(livePrice float64) string {
	var diffTargetPrice = a.TargetPrice - livePrice
	var diffStartPrice = livePrice - a.StartPrice
	var activeIcon string
	if a.Active {
		activeIcon = "\U0001F7E2"
	} else {
		activeIcon = "\U0001F534"
	}
	var diffStartPriceIcon string
	if diffStartPrice == 0 {
		diffStartPriceIcon = "\u27A1\uFE0F"
	} else if diffStartPrice > 0 {
		diffStartPriceIcon = "\u2B06\uFE0F"
	} else {
		diffStartPriceIcon = "\u2B07\uFE0F"
	}
	var diffTargetPriceIcon string
	if diffTargetPrice > 0 {
		diffTargetPriceIcon = "\U0001F538"
	} else {
		diffTargetPriceIcon = "\U0001F539"
	}
	return fmt.Sprintf("#%d [%s] %s %s\n(%.5f) => [%s %.5f]\n(%.5f) => [%s %.5f]",
		a.Number, strings.ToUpper(a.Symbol), activeIcon, a.Description, a.TargetPrice, diffTargetPriceIcon, math.Abs(diffTargetPrice), livePrice, diffStartPriceIcon, diffStartPrice)
}
