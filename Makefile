.PHONY: run

run-bot-local:
	docker-compose up -d tarantool
	go mod tidy
	go run .

run-bot-docker:
	docker-compose up --build