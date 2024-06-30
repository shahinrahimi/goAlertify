package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Printf("GoAlertify Version: %s\n", version)
	if err := godotenv.Load(); err != nil {
		log.Panic("Error loading .env file", err)
	}
	store, err := NewSqliteStore()
	if err != nil {
		log.Panic("Database not found", err)
	}
	if err := store.Init(); err != nil {
		log.Panic("Database does not initialized", err)
	}

	apiKey := os.Getenv("TELEGRAM_BOT_API_KEY")
	if apiKey == "" {
		log.Panic("Telegram bot apiKey not found")
	}

	bot, err := NewTelegramBot(store, apiKey)
	if err != nil {
		log.Panic("Telegram bot does not initialized", err)
	}

	// start scrapper
	go StartScrapping()

	go bot.Run()

	// Keep the main function alive
	log.Println("Start listening for updates.")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Shutting down gracefully...")
}
