# syntax=docker/dockerfile:1
ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -o /out/uptime ./cmd/uptime

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/uptime /app/uptime

USER app

ENTRYPOINT ["/app/uptime"]