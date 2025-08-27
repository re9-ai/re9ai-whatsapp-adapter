# Use an official Go runtime as a parent image
FROM golang:1.21 AS builder

# Set the working directory in the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o re9ai-whatsapp-adapter ./cmd/re9ai-whatsapp-adapter

# Use a smaller base image for the final runtime
FROM alpine:latest

# Set the working directory in the container
WORKDIR /app

# Copy the built application from the builder stage
COPY --from=builder /app/re9ai-whatsapp-adapter .

# Expose the port the app runs on
EXPOSE 8080

# Run the application
CMD ["./re9ai-whatsapp-adapter"]