# Multi-stage build with multi-architecture support
FROM --platform=$TARGETPLATFORM golang:1.22-alpine AS builder

# Build arguments for multi-arch support
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Set the working directory inside the container
WORKDIR /app

# Print build information for debugging
RUN echo "Building on $BUILDPLATFORM for $TARGETPLATFORM (OS: $TARGETOS, ARCH: $TARGETARCH)"

# Copy the source code into the container
COPY . .

# Initialize the module and download dependencies
RUN go mod init trace-generator && \
    go mod tidy

# Build the application with cross-compilation support
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o trace-generator

# Use a slim alpine image for the final stage with multi-arch support
FROM --platform=$TARGETPLATFORM alpine:latest

# Install CA certificates for secure connections
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the built executable from the builder stage
COPY --from=builder /app/trace-generator .

# Set default environment variable (can be overridden at runtime)
ENV JAEGER_ENDPOINT=localhost:4317

# Command to run the executable
ENTRYPOINT ["./trace-generator"]
