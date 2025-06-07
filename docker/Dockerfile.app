# Stage 1: Build the application
FROM golang:1.24.2 AS builder

# Set working directory
WORKDIR /app

# Copy the source code
COPY ./nhp-app .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gin-hello-world

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Install CA certificates (needed for HTTPS requests)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/gin-hello-world .

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["./gin-hello-world"]