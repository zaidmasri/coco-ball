# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

# go-sqlite3 uses cgo, so we need a C toolchain
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/northbasis-cli ./cmd/cli

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/northbasis-cli /app/northbasis-cli

RUN mkdir -p /app/data

EXPOSE 8080

ENTRYPOINT ["/app/northbasis-cli"]
CMD ["serve", "--db", "/app/data/northbasis.db", "--port", ":8080"]
