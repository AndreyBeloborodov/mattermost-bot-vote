.PHONY: run

run:
	docker-compose up -d tarantool
	go mod tidy
	go run .
