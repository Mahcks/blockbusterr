# Stage 1: Build the Golang backend with static linking
FROM golang:1.22 AS server-builder

WORKDIR /app/server

COPY apps/server/go.mod apps/server/go.sum ./
RUN go mod download

COPY apps/server/ .

WORKDIR /app/server/cmd/app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /app/server/server .

# Stage 2: Build the Vite React frontend
FROM node:20-alpine AS client-builder

WORKDIR /app/client

COPY apps/client/ .

# Pass environment variables at build time
ARG VITE_API_URL

ENV VITE_API_URL='api/v1'

RUN npm install && npm run build

# Stage 3: Final stage, run the server and client together with nginx
FROM nginx:alpine

WORKDIR /app

# Copy the built server
COPY --from=server-builder /app/server/server /app/server/server

# Copy the built client
COPY --from=client-builder /app/client/dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Run the Go server in the background and nginx in the foreground
CMD /app/server/server & nginx -g 'daemon off;'
