FROM golang:1.27.1-trixie AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./gen gen/
COPY ./cmd cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o chat_backend \
    ./cmd/server

FROM alpine:3.24 AS server

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /app/chat_backend .

CMD [ "/app/chat_backend" ]
