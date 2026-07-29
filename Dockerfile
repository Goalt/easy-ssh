# Build Stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Target OS and Architecture are injected automatically by Docker Buildx
ARG TARGETOS
ARG TARGETARCH

# Build statically-linked binary
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o easy-ssh main.go

# Final Runner Stage
FROM alpine:latest

# Install CA certificates (required for secure TLS/Cloudflare Tunnel connections), SSH client, and curl
RUN apk add --no-cache ca-certificates openssh-client curl

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/easy-ssh /usr/local/bin/easy-ssh

# Set default entrypoint
ENTRYPOINT ["/usr/local/bin/easy-ssh"]
