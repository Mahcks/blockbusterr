# Stage 1: Build the Golang backend with CGO enabled
FROM golang:1.22 AS server-builder

# Version of the server build
ARG VERSION=""

WORKDIR /app/server

# Install necessary build tools and libraries
RUN apt-get update && apt-get install -y gcc musl-dev

# Copy go.mod and go.sum
COPY apps/server/go.mod apps/server/go.sum ./
RUN go mod download

# Copy the server source code
COPY apps/server/ .

# Build the Go server binary with CGO enabled
WORKDIR /app/server/cmd/app
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /app/server/server -ldflags="-X 'main.Version=${VERSION}'" .

# Stage 2: Build the Vite React frontend
FROM node:20-alpine AS client-builder

WORKDIR /app/client

# Copy the client source code
COPY apps/client/ .

# Pass environment variables at build time
ARG VITE_API_URL
ENV VITE_API_URL='v1'

# Install dependencies and build the client
RUN npm install && npm run build

# Stage 3: Final stage, run the server and client together with SQLite and nginx
# Use Debian or Ubuntu to ensure all dynamic libraries are present
FROM debian:bookworm-slim

# Install necessary runtime libraries for Go, Nginx, and CA certificates
RUN apt-get update && apt-get install -y \
    nginx \
    sqlite3 \
    libc6-dev \
    gcc \
    ca-certificates \
    --no-install-recommends && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the built server from the server-builder stage
COPY --from=server-builder /app/server/server /app/server/server

# Copy the built client from the client-builder stage
COPY --from=client-builder /app/client/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Create a data directory for SQLite and copy migrations (if any)
RUN mkdir -p /app/data
COPY ./apps/server/internal/db/migrations /migrations

# Copy the SQLite initialization script
COPY init-db.sh /app/init-db.sh
RUN chmod +x /app/init-db.sh

# Expose required ports
EXPOSE 5555

# Run the SQLite initialization script, Go server, and nginx in the background
ENTRYPOINT ["/app/init-db.sh"]
CMD /app/server/server & nginx -g 'daemon off;'
