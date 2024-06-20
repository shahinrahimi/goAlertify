package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Panic("Error loading .env file")
	}
	store, err := NewSqliteStore()
	if err != nil {
		log.Panic("Database not found")
	}
	if err := store.Init(); err != nil {
		log.Panic("Database does not initialized")
	}

	apiKey := os.Getenv("TELEGRAM_BOT_API_KEY")
	if apiKey == "" {
		log.Panic("Telegram bot apiKey not found")
	}

	bot, err := NewTelegramBot(store, apiKey)
	if err != nil {
		log.Panic("Telegram bot does not initialized")
	}

	bot.Run()
}
