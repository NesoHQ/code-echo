# === Stage 1: Build the binary ===
FROM golang:1.25 AS builder

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire code
COPY . .

# Build the binary (static binary for Linux)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o codeecho ./main.go


# === Stage 2: Create minimal runtime image ===
FROM alpine:3.20

# Create a non-root user for security
RUN adduser -D codeecho
USER codeecho

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/codeecho /usr/local/bin/codeecho

# Default command (prints help)
ENTRYPOINT ["codeecho"]
CMD ["--help"]
