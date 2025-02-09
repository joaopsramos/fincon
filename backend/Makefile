.DEFAULT_GOAL := build

.PHONY:fmt vet build run setup-db test test-cover linter

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o fincon ./cmd/fincon

run:
	go run ./cmd/fincon

setup-db:
	go run ./cmd/setup_db

migrate:
	go run ./cmd/migrate_db

test:
	APP_ENV=test go test ./...

test-cover:
	APP_ENV=test go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out

linter:
	golangci-lint run
