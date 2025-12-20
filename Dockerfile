# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" -o /app/bin/server ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -g '' app

WORKDIR /app

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/templates /app/templates

ENV APP_ENV=production \
    EMAIL_TEMPLATE_DIR=/app/templates/email

EXPOSE 3000

USER app

ENTRYPOINT ["/app/server"]
