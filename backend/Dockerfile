FROM golang:1.23.4

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .

RUN go build -o fincon ./cmd/fincon
RUN go build -o fincon_setup_db ./cmd/setup_db

EXPOSE 4000

CMD ["./fincon"]
