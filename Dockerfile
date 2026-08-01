# syntax=docker/dockerfile:1

FROM golang:1.26-bookworm AS builder

# go-sqlite3 uses cgo; build against glibc to avoid musl-related cgo issues
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/northbasis-cli ./cmd/cli

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bin/northbasis-cli /app/northbasis-cli

RUN mkdir -p /app/data

EXPOSE 8080

ENTRYPOINT ["/app/northbasis-cli"]
CMD ["serve", "--db", "/app/data/northbasis.db", "--port", ":8080"]
