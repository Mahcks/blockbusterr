# Choose Go version
ARG GOLANG_TAG=1.24-alpine

FROM golang:${GOLANG_TAG} AS builder

ARG VERSION=""
ARG COMMIT=""

# Install required tools
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o blockbusterr \
    -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Timestamp=${COMMIT}'" \
    ./cmd/app/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/blockbusterr .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy config example
COPY config/config.example.yaml ./config/config.example.yaml

# Copy web templates and static files
COPY web/ ./web/

# Create data directory
RUN mkdir -p /app/data

# Expose port (hardcoded to 9090)
EXPOSE 9090

# Run the application
CMD ["./blockbusterr"]
