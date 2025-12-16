# Use official Golang image for building
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with static linking (good for alpine runtime)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/server  # Adjust path if your main is elsewhere

# Final stage: lightweight runtime
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/app .

# Make it executable (this fixes the permission issue!)
RUN chmod +x app

# Expose port (match your config.AppPort(), usually 8080)
EXPOSE 8080

# Run the app
CMD ["./app"]