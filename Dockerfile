FROM golang:1.26-alpine AS builder
RUN apk add --no-cache build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o event-relay ./cmd/app/main.go

FROM alpine:latest
RUN apk add --no-cache libc6-compat
WORKDIR /var/lib/outbox
COPY --from=builder /app/event-relay .
RUN chown -R nobody:nobody /var/lib/outbox
USER nobody
ENTRYPOINT ["./event-relay"]