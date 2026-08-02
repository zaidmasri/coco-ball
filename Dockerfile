# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine3.24 AS builder

# go-sqlite3 uses cgo, so we need a C toolchain
RUN apk add --no-cache gcc=15.2.0-r5 musl-dev=1.2.6-r2

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o /app/bin/northbasis-cli ./cmd/cli

FROM alpine:3.24

RUN apk add --no-cache ca-certificates=20260611-r0

WORKDIR /app

COPY --from=builder /app/bin/northbasis-cli /app/northbasis-cli

RUN mkdir -p /app/data

EXPOSE 8080

ENTRYPOINT ["/app/northbasis-cli"]
CMD ["serve", "--db", "/app/data/northbasis.db", "--port", ":8080"]
