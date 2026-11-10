FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN apk add --no-cache git wget
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/auth ./cmd/app
RUN CGO_ENABLED=0 go build -o /app/migrate ./cmd/migrate

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/auth ./auth
COPY --from=builder /app/migrate ./migrate

EXPOSE 50150

CMD ["./auth"]