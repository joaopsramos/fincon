.DEFAULT_GOAL := build

.PHONY:fmt vet build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o fincon ./cmd/fincon

run:
	go run ./cmd/fincon

test:
	APP_ENV=test go test
