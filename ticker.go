package main

import (
	"time"
)

type Ticker struct {
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	LivePrice float64   `json:"live_price"`
	DailyHigh float64   `json:"daily_high"`
	DailyLow  float64   `json:"daily_low"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewTicker(symbol, name, category string, livePrice, dailyHigh, dailyLow float64) *Ticker {
	return &Ticker{
		Symbol:    symbol,
		Name:      name,
		Category:  category,
		LivePrice: livePrice,
		DailyHigh: dailyHigh,
		DailyLow:  dailyLow,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func (t *Ticker) Update(livePrice, dailyHigh, dailyLow float64) error {
	t.UpdatedAt = time.Now().UTC()
	t.LivePrice = livePrice
	t.DailyHigh = dailyHigh
	t.DailyLow = dailyLow
	return nil
}
