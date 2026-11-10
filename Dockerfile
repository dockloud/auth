FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN apk add --no-cache git
RUN go mod download

COPY --from=ghcr.io/grpc-ecosystem/grpc-health-probe:v0.4.61 /ko-app/grpc-health-probe /bin/grpc_health_probe

COPY . .

RUN CGO_ENABLED=0 go build -o /app/auth ./cmd/app
RUN CGO_ENABLED=0 go build -o /app/migrate ./cmd/migrate

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/auth ./auth
COPY --from=builder /app/migrate ./migrate
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe

EXPOSE 50150

CMD ["./auth"]