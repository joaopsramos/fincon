FROM golang:1.23.4

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .

ENV APP_ENV="prod"

RUN go build -o fincon ./cmd/fincon
RUN go build -o fincon_migrate_db ./cmd/migrate_db

EXPOSE 4000

CMD ["./fincon"]
