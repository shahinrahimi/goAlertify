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

	switch {
	case strings.HasPrefix(command, "/start"):
		err = b.registerUser(chatId, userId, username, fisrtname, lastname)
	case strings.HasPrefix(command, "/viewuser"):
		err = b.viewUser(chatId, userId)
	case strings.HasPrefix(command, "/deleteuser"):
		err = b.deleteUser(chatId, userId)
	case strings.HasPrefix(command, "/createalert"):
		err = b.createAlert(chatId, userId, command)
	case strings.HasPrefix(command, "/viewalerts"):
		err = b.viewAlerts(chatId, userId, command)
	case strings.HasPrefix(command, "/updatealert"):
		err = b.updateAlert(chatId, userId, command)
	case strings.HasPrefix(command, "/deletealert"):
		err = b.deleteAlert(chatId, userId, command)
	case strings.HasPrefix(command, "/viewsymbols"):
		err = b.viewSymbols(chatId, userId)
	default:
		// Handle unknown commands or provide instructions
		msg := tgbotapi.NewMessage(chatId, "Unknown command. Available commands: /start, /createalert, /updatealert, /deletealert, /viewalerts, /viewsymbols")
		_, err := b.bot.Send(msg)
		if err != nil {
			log.Println("Error sending message:", err)
		}
	}

	return err
}

// crud on user
func (b *TelegramBot) registerUser(chatId, userId int64, username, fisrtname, lastname string) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user != nil {
		msg := tgbotapi.NewMessage(chatId, "You are already registered.")
		_, err := b.bot.Send(msg)
		return err
	}

	newUser, err := NewUser(userId, username, fisrtname, lastname, "default_pass")
	if err != nil {
		msg := tgbotapi.NewMessage(chatId, "Error creating user object.")
		_, err := b.bot.Send(msg)
		return err
	}

	if err := b.store.CreateUser(*newUser); err != nil {
		fmt.Println(err)
		msg := tgbotapi.NewMessage(chatId, "Error storing user to DB.")
		_, err := b.bot.Send(msg)
		return err
	}
	msg := tgbotapi.NewMessage(chatId, "You have been registered successfully.")
	_, err = b.bot.Send(msg)
	return err
}
func (b *TelegramBot) viewUser(chatId, userId int64) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil {
		if err == sql.ErrNoRows {
			msg := tgbotapi.NewMessage(chatId, "User not found.")
			_, err := b.bot.Send(msg)
			return err
		}
		return err
	}

	// Format user data for display
	userData := fmt.Sprintf("User ID: %d\nChat ID: %d\nUsername: %s\nFistname: %s\nLastname: %s\nPassword: %s\nCreated At: %s",
		user.UserId, chatId, user.Username, user.Firstname, user.Lastname, user.Password, user.CreatedAt.Format(time.RFC3339))

	msg := tgbotapi.NewMessage(chatId, userData)
	_, err = b.bot.Send(msg)
	return err
}
func (b *TelegramBot) deleteUser(chatId, userId int64) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	if err := b.store.DeleteUserAndAlerts(userId); err != nil {
		msg := tgbotapi.NewMessage(chatId, "user data can not be deleted.")
		_, _ = b.bot.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(chatId, "User and all associated alerts have been deleted successfully.")
	_, err = b.bot.Send(msg)
	return err
}

// crud alert
func (b *TelegramBot) createAlert(chatId, userId int64, command string) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	// Extract alert details from the command
	parts := strings.SplitN(command, " ", 4)
	println(len(parts))
	if len(parts) < 3 {
		msg := tgbotapi.NewMessage(chatId, "Usage: /createalert <ticker> <traget_price> <description>")
		_, err := b.bot.Send(msg)
		return err
	}
	tickerSymbol := strings.ToUpper(parts[1])
	t, exist := tickers[tickerSymbol]
	if !exist {
		msg := tgbotapi.NewMessage(chatId, "Symbol not found, please try later or insert valid symbol.")
		_, err := b.bot.Send(msg)
		return err
	}

	targetPrice, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		msg := tgbotapi.NewMessage(chatId, "Invalid target price.")
		_, err = b.bot.Send(msg)
		return err
	}

	// check if target_price is not in the range of daily (High and Low)
	if targetPrice < t.DailyHigh && targetPrice > t.DailyLow {
		msg := tgbotapi.NewMessage(chatId, "Invalid target price.\ntarget_price already in range of daily hight and low.")
		_, err = b.bot.Send(msg)
		return err
	}
	var description string
	if len(parts) == 4 {
		description = parts[3]
	} else {
		description = ""
	}
	newAlert := NewAlert(userId, t.Symbol, description, targetPrice, t.LivePrice)
	if err := b.store.CreateAlert(newAlert); err != nil {
		msg := tgbotapi.NewMessage(chatId, "Error storing the alert.")
		_, err = b.bot.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(chatId, "Alert added successfully.")
	_, err = b.bot.Send(msg)
	return err
}
func (b *TelegramBot) viewAlerts(chatId, userId int64, command string) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	var alerts []Alert
	parts := strings.SplitN(command, " ", 2)
	if len(parts) == 1 {
		// view all alerts
		alerts, err = b.store.GetAlertsByUserId(userId)
		log.Println(len(alerts))
		if err != nil {
			return err
		}
	} else if len(parts) == 2 {
		// view alerts with symbol
		symbol := parts[1]
		alerts, err = b.store.GetAlertsByUserIdAndSymbol(userId, symbol)
		if err != nil {
			return err
		}
	} else {
		msg := tgbotapi.NewMessage(chatId, "Usage: /viewalerts <symbol(optional)>")
		_, err := b.bot.Send(msg)
		return err
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
		msg := tgbotapi.NewMessage(chatId, "No alerts found.")
		_, err := b.bot.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(chatId, strings.Join(alertStrings, "\n\n"))
	_, err = b.bot.Send(msg)
	return err
}
func (b *TelegramBot) updateAlert(chatId, userId int64, command string) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	// Extract alert details from the command
	parts := strings.SplitN(command, " ", 3)
	if len(parts) < 3 {
		msg := tgbotapi.NewMessage(chatId, "Usage: /updatealert <number> <target_price>")
		_, err := b.bot.Send(msg)
		return err
	}

	number, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(chatId, "Invalid alert number.")
		_, err = b.bot.Send(msg)
		return err
	}
	alert, err := b.store.GetAlertByNumber(userId, int32(number))
	if err != nil {
		msg := tgbotapi.NewMessage(chatId, "Alert not found.")
		_, err := b.bot.Send(msg)
		return err
	}

	targetPrice, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		msg := tgbotapi.NewMessage(chatId, "Invalid target price.")
		_, err = b.bot.Send(msg)
		return err
	}
	ticker, exists := tickers[alert.Symbol]
	if !exists {
		msg := tgbotapi.NewMessage(chatId, "Live price not available for editing alert")
		_, err := b.bot.Send(msg)
		return err
	}

	alert.StartPrice = ticker.LivePrice
	alert.TargetPrice = targetPrice
	if err := b.store.UpdateAlert(alert); err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatId, "Alert updated successfully.")
	_, err = b.bot.Send(msg)
	return err
}
func (b *TelegramBot) deleteAlert(chatId, userId int64, command string) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	parts := strings.SplitN(command, " ", 2)
	if len(parts) < 2 {
		msg := tgbotapi.NewMessage(chatId, "Usage: /deletealert <number>")
		_, err := b.bot.Send(msg)
		return err
	}

	number, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(chatId, "Invalid alert number.")
		_, err = b.bot.Send(msg)
		return err
	}

	alert, err := b.store.GetAlertByNumber(userId, int32(number))
	if err != nil {
		msg := tgbotapi.NewMessage(chatId, "Alert not found.")
		_, err := b.bot.Send(msg)
		return err
	}

	if err := b.store.DeleteAlert(alert.Id); err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatId, "Alert deleted successfully.")
	_, err = b.bot.Send(msg)
	return err
}

func (b *TelegramBot) viewSymbols(chatId, userId int64) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	var tickerStrings []string
	for _, ticker := range tickers {
		tickerStrings = append(tickerStrings, fmt.Sprintf("Symbol: %s\n\tNow: %.4f\n\tHigh: %.4f\n\tLow: %.4f\n\n",
			ticker.Symbol, ticker.LivePrice, ticker.DailyHigh, ticker.DailyLow))
	}

	if len(tickerStrings) == 0 {
		msg := tgbotapi.NewMessage(chatId, "No tickers found.")
		_, err := b.bot.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(chatId, strings.Join(tickerStrings, "\n\n"))
	_, err = b.bot.Send(msg)
	return err
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
