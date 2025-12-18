#!/bin/bash

# Build script for cache-engine

echo "Building cache-engine..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build the application
go build -o bin/cache-engine ./cmd/cache-engine

if [ $? -eq 0 ]; then
    echo "✓ Build successful!"
    echo "Binary created at: bin/cache-engine"
    echo ""
    echo "Run with: ./bin/cache-engine -mode=cli"
    echo "       or: ./bin/cache-engine -mode=http -port=8080"
else
    echo "✗ Build failed"
    exit 1
fi