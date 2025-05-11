# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bot ./cmd/bot

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/bot /app/bot

# Create directory for migrations
RUN mkdir -p /app/internal/infrastructure/database/migrations

# Copy migrations
COPY --from=builder /app/internal/infrastructure/database/migrations /app/internal/infrastructure/database/migrations

# Set the timezone
ENV TZ=UTC

# Run as non-root user
RUN adduser -D appuser
USER appuser

# Command to run the application
CMD ["/app/bot"]