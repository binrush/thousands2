# Use Debian 12 as base image
FROM debian:12

ENV GOVERSION=1.25.0

# Install dependencies: Go, Node.js 18, curl, and build essentials
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    ca-certificates \
    gnupg \
    && curl -fsSL https://deb.nodesource.com/setup_18.x | bash - \
    && apt-get install -y nodejs \
    && curl -LO https://go.dev/dl/go${GOVERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GOVERSION}.linux-amd64.tar.gz \
    && rm go${GOVERSION}.linux-amd64.tar.gz \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/local/go/bin:$PATH"

WORKDIR /app

# Copy the entire project
COPY . .

# Build the UI
RUN cd src/ui && npm ci && npm run build

# Build the Go binary
RUN cd src && \
    CGO_ENABLED=1 \
    go build -ldflags="-s -w" -o thousands2

# Create output directory
RUN mkdir -p /dist

# Copy the binary to the output directory
RUN cp src/thousands2 /dist/thousands2-linux-amd64 