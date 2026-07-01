# ---------- Builder ----------
FROM golang:1.25.3 AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service ./cmd/main.go


# ---------- Runtime ----------
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/payment-service .

# Payment Service Port
EXPOSE 8085

# Run service
CMD ["./payment-service"]
