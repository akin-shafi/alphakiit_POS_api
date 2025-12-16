# Use official Golang image for building
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files first (better layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the real server binary from cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o app ./cmd/server/main.go

# Final lightweight runtime
FROM alpine:latest

WORKDIR /app

# Copy only the compiled binary
COPY --from=builder /app/app .

# Make it executable
RUN chmod +x app

# Expose port (your app listens on 8080 or from config)
EXPOSE 8080

# Run the app
CMD ["./app"]