FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN GOOS=linux GOARCH=arm64 go build -o fan-controller main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    kmod \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app/fan-controller .

ENTRYPOINT ["./fan-controller"]
