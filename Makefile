build:
	@go build -o ./bin/golalert

run: build
	@./bin/golalert