# Choose Go version
ARG GOLANG_TAG=1.25.12-alpine

FROM golang:${GOLANG_TAG} AS builder

ARG VERSION=""
ARG COMMIT=""

# Install required tools (including gcc and musl-dev for CGO/SQLite)
# Use --no-scripts to avoid trigger issues with QEMU emulation in multi-arch builds
RUN apk add --no-cache --no-scripts ca-certificates git gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version info (CGO_ENABLED=1 for SQLite)
RUN CGO_ENABLED=1 GOOS=linux go build \
    -o blockbusterr \
    -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}'" \
    ./cmd/app/main.go

# Runtime stage
FROM alpine:3.22

# Install ca-certificates for HTTPS requests
# Use --no-scripts to avoid trigger issues with QEMU emulation in multi-arch builds
RUN apk --no-cache --no-scripts add ca-certificates tzdata && \
    update-ca-certificates && \
    addgroup -g 10001 blockbusterr && \
    adduser -D -u 10001 -G blockbusterr blockbusterr

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/blockbusterr .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy config example
COPY config/config.example.yaml ./config/config.example.yaml

# Copy web templates and static files
COPY web/ ./web/

# Create data directory
RUN mkdir -p /app/data && chown -R blockbusterr:blockbusterr /app

# Expose port (hardcoded to 9090)
EXPOSE 9090

USER 10001:10001

# Run the application
CMD ["./blockbusterr"]
