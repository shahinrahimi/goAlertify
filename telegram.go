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
	ctx, cancel := context.WithCancel(ctx)

	updates := b.bot.GetUpdatesChan(u)

	go b.receiveUpdates(ctx, updates)
	go b.startAlertChecker()

	log.Println("Start listening for updates. Press enter to stop")

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
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
	case strings.HasPrefix(command, "/viewtickers"):
		err = b.viewTickers(chatId, userId)
	case strings.HasPrefix(command, "/menu"):
		err = b.sendMainMenu(chatId)
	case strings.HasPrefix(command, "/addalert"):
		err = b.addAlert(chatId, userId, command)
	case strings.HasPrefix(command, "/viewalerts"):
		err = b.viewAlerts(chatId, userId)
	case strings.HasPrefix(command, "/updatealert"):
		err = b.updateAlert(chatId, userId, command)
	case strings.HasPrefix(command, "/deletealert"):
		err = b.deleteAlert(chatId, userId, command)
	case strings.HasPrefix(command, "/viewuser"):
		err = b.viewUser(chatId, userId)
	case strings.HasPrefix(command, "/deleteuser"):
		err = b.deleteUser(chatId, userId)
	}

	return err
}

// crud on user
func (b *TelegramBot) registerUser(chatId int64, userId int64, username, fisrtname, lastname string) error {
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
func (b *TelegramBot) viewUser(chatId int64, userId int64) error {
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
func (b *TelegramBot) deleteUser(chatId int64, userId int64) error {
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
func (b *TelegramBot) addAlert(chatId int64, userId int64, command string) error {
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
	if len(parts) < 4 {
		msg := tgbotapi.NewMessage(chatId, "Usage: /addalert <ticker> <traget_price> <description>")
		_, err := b.bot.Send(msg)
		return err
	}

	t, exist := tickers[parts[1]]
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

	description := parts[3]
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
func (b *TelegramBot) viewAlerts(chatId int64, userId int64) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	alerts, err := b.store.GetAlertsByUserId(userId)
	if err != nil {
		return err
	}

	var alertStrings []string
	for _, alert := range alerts {
		t, _ := tickers[alert.Symbol]
		alertStrings = append(alertStrings, fmt.Sprintf("ID: %s\nSymbol: %s\nFrom: %.8f ==> %.8f\nCurr: %.8f\nActive: %t, \nCreated At: %s",
			alert.Id, alert.Symbol, alert.StartPrice, alert.TargetPrice, t.LivePrice, alert.Active, alert.CreatedAt))
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
func (b *TelegramBot) updateAlert(chatId int64, userId int64, command string) error {
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
	if len(parts) < 3 {
		msg := tgbotapi.NewMessage(chatId, "Usage: /updatealert <id> <target_price>")
		_, err := b.bot.Send(msg)
		return err
	}

	id := parts[1]
	alerts, err := b.store.GetAlertsByUserId(userId)
	isUserHasPermision := ContainsAlert(alerts, id)
	if !isUserHasPermision {
		msg := tgbotapi.NewMessage(chatId, "The alert not found")
		_, err := b.bot.Send(msg)
		return err
	}

	alert, err := b.store.GetAlert(id)
	if err != nil {
		msg := tgbotapi.NewMessage(chatId, "The alert not found")
		_, err := b.bot.Send(msg)
		return err
	}
	targetPrice, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return err
	}
	t, exists := tickers[alert.Symbol]
	if !exists {
		msg := tgbotapi.NewMessage(chatId, "Live price not available for editing alert")
		_, err := b.bot.Send(msg)
		return err
	}
	alert.StartPrice = t.LivePrice
	alert.TargetPrice = targetPrice
	if err := b.store.UpdateAlert(alert); err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatId, "Alert updated successfully.")
	_, err = b.bot.Send(msg)
	return err
}
func (b *TelegramBot) deleteAlert(chatId int64, userId int64, command string) error {
	user, err := b.store.GetUserByUserId(userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if user == nil {
		msg := tgbotapi.NewMessage(chatId, "You are not registered.\nUsage: /start")
		_, err := b.bot.Send(msg)
		return err
	}
	// Extract alert id from the command
	parts := strings.SplitN(command, " ", 2)
	if len(parts) < 2 {
		msg := tgbotapi.NewMessage(chatId, "Usage: /deletealert <id>")
		_, err := b.bot.Send(msg)
		return err
	}
	id := parts[1]

	// Get user alerts
	alerts, err := b.store.GetAlertsByUserId(userId)
	if err != nil {
		return err
	}

	if isOwner := ContainsAlert(alerts, id); !isOwner {
		msg := tgbotapi.NewMessage(chatId, "Permission Denied!")
		_, err := b.bot.Send(msg)
		return err
	}

	// Delete the alert from the database
	if err := b.store.DeleteAlert(id); err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatId, "Alert deleted successfully.")
	_, err = b.bot.Send(msg)
	return err
}

func (b *TelegramBot) viewTickers(chatId int64, userId int64) error {
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
func (b *TelegramBot) sendMainMenu(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = mainMenuMarkup
	_, err := b.bot.Send(msg)
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

	for _, alert := range alerts {
		if !alert.Active {
			continue
		}
		ticker, exist := tickers[alert.Symbol]
		if !exist {
			fmt.Println("Ticker not found for alert ticker:", alert.Symbol, "id:", alert.Id)
			msg := tgbotapi.NewMessage(alert.UserId, "Ticker not found for alert ticker:"+alert.Symbol)
			_, err = b.bot.Send(msg)
			continue
		}

		if ticker.LivePrice == alert.TargetPrice {
			fmt.Println("Alert hit pirce")
			alert.Active = false
			b.store.UpdateAlert(&alert)
			msg := tgbotapi.NewMessage(alert.UserId, fmt.Sprintf("Alert triggered for %s! Current price: %.8f TargetPrice was: %.8f", alert.Symbol, ticker.LivePrice, alert.TargetPrice))
			_, err := b.bot.Send(msg)
			if err != nil {
				log.Printf("Error sending alert notification to user %d: %s", alert.UserId, err.Error())
			}
			return
		}
		// target_price is higher than start_price => ceil
		// dailyHigh bigger than target_price
		if alert.TargetPrice > alert.StartPrice && ((alert.TargetPrice < ticker.DailyHigh) || (alert.TargetPrice < ticker.LivePrice)) {
			alert.Active = false
			b.store.UpdateAlert(&alert)
			msg := tgbotapi.NewMessage(alert.UserId, fmt.Sprintf("Alert triggered for %s! Current price: %.5f TargetPrice was: %.5f", alert.Symbol, ticker.LivePrice, alert.TargetPrice))
			_, err := b.bot.Send(msg)
			if err != nil {
				log.Printf("Error sending alert notification to user %d: %s", alert.UserId, err.Error())
			}
			return
		}
		// target_price is less than start_price => floor
		// dailyLow less than target_price
		if alert.TargetPrice < alert.StartPrice && ((alert.TargetPrice > ticker.DailyLow) || (alert.TargetPrice > ticker.LivePrice)) {
			alert.Active = false
			b.store.UpdateAlert(&alert)
			msg := tgbotapi.NewMessage(alert.UserId, fmt.Sprintf("Alert triggered for %s! Current price: %.8f TargetPrice was: %.8f", alert.Symbol, ticker.LivePrice, alert.TargetPrice))
			_, err := b.bot.Send(msg)
			if err != nil {
				log.Printf("Error sending alert notification to user %d: %s", alert.UserId, err.Error())
			}
			return
		}
	}
}
