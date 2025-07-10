# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY backend/go-impl/go.mod backend/go-impl/go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY backend/go-impl/ .

RUN go generate ./...

RUN go mod tidy

# Build the application
RUN go build -o main ./cmd/app

# Final stage
FROM alpine:latest

WORKDIR /app

# Create web directory for frontend files
RUN mkdir -p /var/www/html

# Copy the binary from builder
COPY --from=builder /app/main .
COPY backend/go-impl/openapi.yaml .

# Copy frontend public files to web directory
COPY frontend/public/ /var/www/html/

# Copy game_data directory to app directory
COPY game_data/ ./game_data/

# Expose port 8765
EXPOSE 8765

# Command to run the executable
CMD ["./main"] 