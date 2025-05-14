#!/bin/bash
set -e

# Directory setup
ROOT_DIR=$(pwd)
SRC_DIR="$ROOT_DIR/src"
UI_DIR="$SRC_DIR/ui"
DIST_DIR="$ROOT_DIR/dist"

# Default values for GOOS and GOARCH
GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}

echo "Building thousands2 with embedded UI using Docker..."
echo "Target OS: $GOOS"
echo "Target Architecture: $GOARCH"

# Create dist directory if it doesn't exist
mkdir -p "$DIST_DIR"

# Step 1: Build the Docker image
echo "Step 1: Building Docker image..."
docker build -t thousands2-builder .

# Step 2: Create a temporary container and extract the binary
echo "Step 2: Extracting binary from Docker container..."
CONTAINER_ID=$(docker create thousands2-builder)
docker cp "$CONTAINER_ID:/dist/thousands2-linux-amd64" "$DIST_DIR/thousands2"
docker rm "$CONTAINER_ID"

# Make the binary executable
chmod +x "$DIST_DIR/thousands2"

echo "Build complete. Binary is at $DIST_DIR/thousands2"
echo "Run with: ./thousands2 <datadir> <db_path>" 