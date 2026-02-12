# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder


# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies (great for caching)
RUN go mod download

# Copy the entire source code
COPY . .

# Build the server binary (from cmd/server package)
# Output binary named "main"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o main ./cmd/server

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install ca-certificates in case your app needs HTTPS (e.g., DB connections)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy only the compiled binary from builder
COPY --from=builder /app/main .

# Make it executable (safety)
RUN chmod +x main

# Expose port (your app listens on this)
EXPOSE 8080

# Run the binary
CMD ["./main"]