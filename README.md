# GoAlertify

GoAlertify is a Telegram bot written in Go that scrapes financial data and alerts users when certain conditions are met. This bot scrapes data for forex, futures, and cryptocurrencies from TradingView and allows users to set alerts based on the live prices of these assets.

## Features

- Scrapes live financial data from TradingView for forex, futures, and cryptocurrencies.
- Users can register, create, view, update, and delete price alerts.
- Admin users can manage all users and their alerts.
- Alerts are checked periodically and notifications are sent when conditions are met.
- Interactive Telegram bot with inline keyboards for easy navigation.

## Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/shahinrahimi/GoAlertify.git
   cd GoAlertify
   ```
2. Create a .env file in the root directory and add your environment variables:
  ```sh
  TELEGRAM_BOT_API_KEY=your-telegram-bot-api-key
  ADMIN_USER_ID=your-telegram-user-id
  ```
3. Build and run the application:
  ```
  make run
  ```

## Usage
1. Start the bot by running the application.
2. Use the /start command in Telegram to register as a user.
3. Use the following commands to interact with the bot:
  - /createalert <ticker> <target_price> <description>: Create a new alert.
  - /viewalerts: View all your alerts.
  - /updatealert <number> <target_price>: Update an existing alert.
  - /deletealert <number>: Delete an alert.
  - /viewsymbols [cryptos|feature|forex]: View available symbols.

## Development
### Project Structure:
  - main.go: Entry point of the application. Initializes the bot and starts the scraper.
  - telegram.go: Contains the logic for the Telegram bot, including command handlers and message processing.
  - scraper.go: Contains the logic for scraping financial data from TradingView.
  - store.go: Contains the logic for interacting with the SQLite database.
### Dependencies:
  - go-telegram-bot-api: Telegram Bot API library for Go.
  - godotenv: Library for loading environment variables from a .env file.
  - goquery: Library for scraping and manipulating HTML.

## Lisence
This project is licensed under the MIT License. See the LICENSE file for details.


