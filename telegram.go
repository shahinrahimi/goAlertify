package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot   *tgbotapi.BotAPI
	store Storage
}

var (
	// Database connection
	// Menu texts
	firstMenu = "<b>Main Menu</b>\n\nChoose an option."
	alertMenu = "<b>Alert Menu</b>\n\n1. Add Alert\n2. View Alerts\n3. Update Alert\n4. Delete Alert\n\nUse /addalert, /viewalerts, /updatealert, /deletealert commands respectively."

	// Button texts
	alertButton = "Manage Alerts"

	// Store bot screaming status
	screaming = false

	// Keyboard layout for the main menu
	mainMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(alertButton, alertButton),
		),
	)
)

func NewTelegramBot(store Storage, apiKey string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		return nil, err
	}
	return &TelegramBot{
		store: store,
		bot:   bot,
	}, nil
}

func (b *TelegramBot) Run() {
	// debug telegram bot
	b.bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	ctx := context.Background()
	ctx = context.WithoutCancel(ctx)

	updates := b.bot.GetUpdatesChan(u)

	go b.receiveUpdates(ctx, updates)
	go b.startAlertChecker()

	log.Println("Start listening for updates.")

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	// cancel()
}

// handlers
func (b *TelegramBot) receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			b.handleUpdate(update)
		}
	}
}
func (b *TelegramBot) handleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		b.handleMessage(update.Message)
	case update.CallbackQuery != nil:
		b.handleButton(update.CallbackQuery)
	}
}
func (b *TelegramBot) handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	log.Printf("id: %d, %s wrote %s", user.ID, user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = b.handleCommand(message.Chat.ID, user.ID, text, user.UserName, user.FirstName, user.LastName)
	} else if screaming && len(text) > 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, strings.ToUpper(text))
		msg.Entities = message.Entities
		_, err = b.bot.Send(msg)
	} else {
		copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
		_, err = b.bot.CopyMessage(copyMsg)
	}

	if err != nil {
		log.Printf("An error occurred: %s", err.Error())
	}
}
func (b *TelegramBot) handleButton(query *tgbotapi.CallbackQuery) {
	var text string
	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	switch query.Data {
	case alertButton:
		text = alertMenu
	default:
		text = firstMenu
		markup = mainMenuMarkup
	}

	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	b.bot.Send(callbackCfg)

	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = tgbotapi.ModeHTML
	b.bot.Send(msg)
}
func (b *TelegramBot) handleCommand(chatId int64, userId int64, command, username, fisrtname, lastname string) error {
	var err error
	var commandParts []string
	parts := strings.Split(command, " ")
	for _, part := range parts {
		commandParts = append(commandParts, strings.ToLower(strings.TrimSpace(part)))
	}
	var mainCommand = commandParts[0]

	switch {
	case mainCommand == "/start":
		err = b.registerUser(chatId, userId, username, fisrtname, lastname)
	case mainCommand == "/viewuser":
		err = b.viewUser(chatId, userId)
	case mainCommand == "/deleteuser":
		err = b.deleteUser(chatId, userId)
	case mainCommand == "/createalert":
		err = b.createAlert(chatId, userId, commandParts[1:])
	case mainCommand == "/viewalerts":
		err = b.viewAlerts(chatId, userId, commandParts[1:])
	case mainCommand == "/updatealert":
		err = b.updateAlert(chatId, userId, commandParts[1:])
	case mainCommand == "/deletealert":
		err = b.deleteAlert(chatId, userId, commandParts[1:])
	case mainCommand == "/viewsymbols":
		err = b.viewSymbols(chatId, userId, commandParts[1:])
	default:
		// Handle unknown commands or provide instructions
		return b.sendMessage(chatId, "Unknown command. Available commands: /start, /createalert, /updatealert, /deletealert, /viewalerts, /viewsymbols")
	}

	return err
}

func (b *TelegramBot) checkUser(userID, chatId int64) (*User, error) {
	user, err := b.store.GetUserByUserId(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, b.sendMessage(chatId, "You are not registered.\nUsage: /start")
		}
		return nil, err
	}
	return user, nil
}

// crud on user
func (b *TelegramBot) registerUser(chatId, userId int64, username, fisrtname, lastname string) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user != nil {
		return b.sendMessage(chatId, "You are already registered.")
	}

	newUser, err := NewUser(userId, username, fisrtname, lastname, "default_pass")
	if err != nil {
		return b.sendMessage(chatId, "Error creating user object.")
	}

	if err := b.store.CreateUser(*newUser); err != nil {
		return b.sendMessage(chatId, "Error storing user to DB.")
	}
	return b.sendMessage(chatId, "You have been registered successfully.")
}
func (b *TelegramBot) viewUser(chatId, userId int64) error {
	user, err := b.checkUser(userId, chatId)
	if err != nil {
		return err
	}
	return b.sendMessage(chatId, user.toTelegramString())
}
func (b *TelegramBot) deleteUser(chatId, userId int64) error {
	user, err := b.checkUser(userId, chatId)
	if err != nil {
		return err
	}
	if err := b.store.DeleteUserAndAlerts(user.UserId); err != nil {
		return b.sendMessage(chatId, "User data can not be deleted.")
	}
	return b.sendMessage(chatId, "User and all associated alerts have been deleted successfully.")
}

// crud alert
func (b *TelegramBot) createAlert(chatId, userId int64, command []string) error {
	_, err := b.checkUser(userId, chatId)
	if err != nil {
		return err
	}

	// Extract alert details from the command
	if len(command) < 2 {
		return b.sendMessage(chatId, "Usage: /createalert <ticker> <traget_price> <description>")
	}

	tickerSymbol := command[0]
	t, exist := tickers[tickerSymbol]
	if !exist {
		return b.sendMessage(chatId, "Symbol not found, please try later or insert valid symbol.")
	}

	targetPrice, err := strconv.ParseFloat(command[1], 64)
	if err != nil {
		return b.sendMessage(chatId, "Invalid target price.")
	}

	// check if target_price is not in the range of daily (High and Low)
	if targetPrice < t.DailyHigh && targetPrice > t.DailyLow {
		return b.sendMessage(chatId, "Invalid target price.\ntarget_price already in range of daily hight and low.")
	}

	var description string
	if len(command) > 2 {
		for index, c := range command {
			if index >= 2 {
				description = description + " " + c
			}
		}
	} else {
		description = ""
	}
	newAlert := NewAlert(userId, t.Symbol, description, targetPrice, t.LivePrice)
	if err := b.store.CreateAlert(newAlert); err != nil {
		return b.sendMessage(chatId, "Error storing the alert.")
	}
	return b.sendMessage(chatId, "Alert added successfully.")
}
func (b *TelegramBot) viewAlerts(chatId, userId int64, command []string) error {
	_, err := b.checkUser(userId, chatId)
	if err != nil {
		return err
	}
	var alerts []Alert
	if len(command) > 0 {
		symbolStr := command[0]
		alerts, err = b.store.GetAlertsByUserIdAndSymbol(userId, symbolStr)
		if err != nil {
			return err
		}
	} else {
		alerts, err = b.store.GetAlertsByUserId(userId)
		if err != nil {
			return err
		}

	}

	var alertStrings []string
	var livePrice float64
	for _, alert := range alerts {
		t, exist := tickers[alert.Symbol]
		if exist {
			livePrice = t.LivePrice
		} else {
			livePrice = 0
		}

		alertStrings = append(alertStrings, alert.ToString(livePrice))
	}

	if len(alertStrings) == 0 {
		return b.sendMessage(chatId, "No alerts found.")
	}
	return b.sendMessageInChunks(chatId, strings.Join(alertStrings, "\n\n"))
}
func (b *TelegramBot) updateAlert(chatId, userId int64, command []string) error {
	_, err := b.checkUser(userId, chatId)
	if err != nil {
		return err
	}

	if len(command) < 2 {
		return b.sendMessage(chatId, "Usage: /updatealert <number> <target_price>")
	}

	number, err := strconv.ParseInt(command[0], 10, 32)
	if err != nil {
		return b.sendMessage(chatId, "Invalid alert number.")
	}
	alert, err := b.store.GetAlertByNumber(userId, int32(number))
	if err != nil {
		return b.sendMessage(chatId, "Alert not found.")
	}

	targetPrice, err := strconv.ParseFloat(command[1], 64)
	if err != nil {
		return b.sendMessage(chatId, "Invalid target price.")
	}
	ticker, exists := tickers[alert.Symbol]
	if !exists {
		return b.sendMessage(chatId, "Live price not available for editing alert")
	}

	alert.StartPrice = ticker.LivePrice
	alert.TargetPrice = targetPrice
	if err := b.store.UpdateAlert(alert); err != nil {
		return err
	}
	return b.sendMessage(chatId, "Alert updated successfully.")
}
func (b *TelegramBot) deleteAlert(chatId, userId int64, command []string) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	if len(command) < 1 {
		return b.sendMessage(chatId, "Usage: /deletealert <number>")
	}

	number, err := strconv.ParseInt(command[0], 10, 32)
	if err != nil {
		return b.sendMessage(chatId, "Invalid alert number.")
	}

	alert, err := b.store.GetAlertByNumber(userId, int32(number))
	if err != nil {
		return b.sendMessage(chatId, "Alert not found.")
	}

	if err := b.store.DeleteAlert(alert.Id); err != nil {
		return err
	}
	return b.sendMessage(chatId, "Alert deleted successfully.")
}

func (b *TelegramBot) sendMessage(chatId int64, msgStr string) error {
	msg := tgbotapi.NewMessage(chatId, msgStr)
	_, err := b.bot.Send(msg)
	return err
}
func (b *TelegramBot) sendMessageInChunks(chatId int64, msgStr string) error {
	const maxMessageSize = 4096
	// Split the message into chunks
	parts := SplitMessage(msgStr, maxMessageSize)
	// Send each chunk
	for _, part := range parts {
		if err := b.sendMessage(chatId, part); err != nil {
			return err
		}
	}
	return nil
}

func (b *TelegramBot) viewSymbols(chatId, userId int64, command []string) error {
	_, err := b.checkUser(chatId, userId)
	if err != nil {
		return err
	}

	var tickerStrings []string
	if len(command) > 0 {
		if command[0] == "cryptos" {
			for _, ticker := range tickers {
				if ticker.Category == "crypto" {
					tickerStrings = append(tickerStrings, ticker.toTelegramString())
				}
			}
		} else if command[0] == "feature" {
			for _, ticker := range tickers {
				if ticker.Category == "feature" {
					tickerStrings = append(tickerStrings, ticker.toTelegramString())
				}
			}

		} else if command[0] == "forex" {
			for _, ticker := range tickers {
				if ticker.Category == "forex" {
					tickerStrings = append(tickerStrings, ticker.toTelegramString())
				}
			}

		} else {
			for _, ticker := range tickers {
				if strings.Contains(ticker.Symbol, command[0]) || strings.Contains(ticker.Name, command[0]) {
					tickerStrings = append(tickerStrings, ticker.toTelegramString())
				}
			}
		}
	} else {
		for _, ticker := range tickers {
			tickerStrings = append(tickerStrings, ticker.toTelegramString())
		}
	}

	if len(tickerStrings) == 0 {
		return b.sendMessage(chatId, "No tickers found.")
	}

	return b.sendMessageInChunks(chatId, strings.Join(tickerStrings, "\n"))
}
func (b *TelegramBot) startAlertChecker() {
	for {
		b.checkAlert()
		time.Sleep(1 * time.Minute)
	}
}
func (b *TelegramBot) checkAlert() {
	alerts, err := b.store.GetAlerts()
	if err != nil {
		log.Printf("Error retrieving alerts: %s", err.Error())
		return
	}

	today := time.Now().Truncate(24 * time.Hour)

	for _, alert := range alerts {
		// reseting alarms if triggered yesterday
		if !alert.Active && alert.UpdatedAt.Before(today) {
			alert.Active = true
			alert.UpdatedAt = time.Now().UTC()
			if err := b.store.UpdateAlert(&alert); err != nil {
				log.Println("Error updating alert", err)
			}
			continue
		}
		ticker, exist := tickers[alert.Symbol]
		if !exist {
			log.Println("Symbol not found:", alert.Symbol, "id:", alert.Id)
			msg := tgbotapi.NewMessage(alert.UserId, "Symbol not found:"+alert.Symbol)
			_, err = b.bot.Send(msg)
			continue
		}

		var isTriggered bool

		if ticker.DailyHigh <= 0 || ticker.DailyLow <= 0 {
			//check alert with etimated fixed amount for example (1% price)
			estimatedAmount := 0.01 * alert.TargetPrice
			if ticker.LivePrice == alert.TargetPrice {
				isTriggered = true
			} else if alert.TargetPrice > alert.StartPrice && alert.TargetPrice < (ticker.LivePrice+estimatedAmount) {
				isTriggered = true
			} else if alert.TargetPrice < alert.StartPrice && alert.TargetPrice > (ticker.LivePrice-estimatedAmount) {
				isTriggered = true
			} else {
				isTriggered = false
			}
		} else {
			//check akert with dailyhigh and dailylow
			if ticker.LivePrice == alert.TargetPrice {
				isTriggered = true
			} else if alert.TargetPrice > alert.StartPrice && ((alert.TargetPrice < ticker.DailyHigh) || (alert.TargetPrice < ticker.LivePrice)) {
				isTriggered = true
			} else if alert.TargetPrice < alert.StartPrice && ((alert.TargetPrice > ticker.DailyLow) || (alert.TargetPrice > ticker.LivePrice)) {
				isTriggered = true
			} else {
				isTriggered = false
			}
		}

		if isTriggered {
			alert.Active = false
			alert.UpdatedAt = time.Now().UTC()
			if err := b.store.UpdateAlert(&alert); err != nil {
				log.Println("Error updating alert", err)
			}
			msg := tgbotapi.NewMessage(alert.UserId, fmt.Sprintf("Alert triggered for %s! Current price: %.5f TargetPrice was: %.5f, with Description: %s", alert.Symbol, ticker.LivePrice, alert.TargetPrice, alert.Description))
			if _, err := b.bot.Send(msg); err != nil {
				log.Printf("Error sending alert notification to user %d: %s", alert.UserId, err.Error())
				return
			}
		}
	}
}
