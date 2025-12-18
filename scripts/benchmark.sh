#!/bin/bash

# Benchmark script for cache-engine

echo "Running benchmarks..."
echo ""

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./internal/cache/

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Benchmarks completed successfully!"
else
    echo ""
    echo "✗ Benchmarks failed"
    exit 1
fi