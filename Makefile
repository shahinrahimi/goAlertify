build:
	@go build -o ./bin/goAlertify

run: build
	@./bin/goAlertify