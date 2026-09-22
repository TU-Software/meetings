FROM golang:1.27.1-trixie AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./db db/
COPY ./gen gen/
COPY ./cmd cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o chat_backend \
    ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o chat_migrate \
    ./cmd/migrate

FROM alpine:3.24 AS server

WORKDIR /app

COPY --from=builder /app/chat_backend .

CMD [ "/app/chat_backend" ]
