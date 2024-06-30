APP_NAME := goAlertify
VERSION := 1.0,1
BUILD_DIR := bin
build:
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-X 'main.version=$(VERSION)'"  -o $(BUILD_DIR)/$(APP_NAME)

run: build
	@./$(BUILD_DIR)/$(APP_NAME)