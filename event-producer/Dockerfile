# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
# These are copied separately to leverage Docker's layer caching.
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
RUN go mod verify

# Copy the source code into the container
COPY . .

# Build the Go app
# The main package for the event-producer is under cmd/producer
# Building a statically-linked binary for a minimal runtime image.
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -installsuffix cgo -o /event-producer ./cmd/producer/main.go

# Stage 2: Create the runtime image from a minimal base image
FROM alpine:latest

# Add a non-root user and group for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the Current Working Directory for the runtime stage
WORKDIR /home/appuser

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /event-producer /event-producer

# Set the user for the runtime image
USER appuser

# Command to run the executable
ENTRYPOINT ["/event-producer"] 