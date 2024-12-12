# Use the official Golang image as the base image
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Initialize the module and download dependencies in one step
RUN go mod init trace-generator && \
    go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o trace-generator

# Use a slim alpine image for the final stage
FROM alpine:latest

# Install CA certificates for secure connections
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the built executable from the builder stage
COPY --from=builder /app/trace-generator .

# Set environment variable (can be overridden at runtime)
ENV JAEGER_ENDPOINT=localhost:4317

# Command to run the executable
ENTRYPOINT ["./trace-generator"]
