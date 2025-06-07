# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker layer caching
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download
RUN go mod verify

# Copy the source code into the container
COPY . .

# Build the Go app
# Building a statically-linked binary for a minimal runtime image
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -installsuffix cgo -o /stream-processor ./cmd/main.go

# Stage 2: Create the runtime image from a minimal base image
FROM alpine:latest

# Install ca-certificates for HTTPS connections and other SSL/TLS certificates
RUN apk --no-cache add ca-certificates tzdata

# Add a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the Current Working Directory for the runtime stage
WORKDIR /home/appuser

# Copy the pre-built binary file from the previous stage
COPY --from=builder /stream-processor /stream-processor

# Create directory for DuckDB if needed
RUN mkdir -p /home/appuser/data && chown -R appuser:appgroup /home/appuser

# Set the user for the runtime image
USER appuser

# Expose any ports if needed (none for this stream processor)

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD pgrep -f stream-processor || exit 1

# Command to run the executable
ENTRYPOINT ["/stream-processor"] 