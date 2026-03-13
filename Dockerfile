# Build stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o zpcli .

# Final stage
FROM alpine:latest

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/zpcli /usr/local/bin/zpcli

# Set the environment variable for the config file
ENV ZPCLI_CONFIG=/etc/zpcli/sites.json

# Expose the SSE port
EXPOSE 8080

# Default command to run the MCP server in SSE mode
ENTRYPOINT ["zpcli"]
CMD ["mcp", "--port", "8080"]
