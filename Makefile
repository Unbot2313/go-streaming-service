.PHONY: build run worker test test-coverage lint swagger docker-build docker-up migrate-diff migrate-apply migrate-status

build:
	go build -o bin/server main.go
	go build -o bin/worker cmd/rabbitmq/consumer/main.go

run:
	go run main.go

worker:
	go run cmd/rabbitmq/consumer/main.go

test:
	go test ./... -race -v

test-coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	go vet ./...

swagger:
	swag init

docker-build:
	docker compose build

docker-up:
	docker compose up --build

migrate-diff:
	atlas migrate diff $(name) --env gorm

migrate-apply:
	atlas migrate apply --env gorm --url "$(DATABASE_URL)"

migrate-status:
	atlas migrate status --env gorm --url "$(DATABASE_URL)"
