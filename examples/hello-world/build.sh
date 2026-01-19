#!/bin/bash
set -e

echo "Building hello-world application..."

# Create bin directory
mkdir -p bin

# Build the binary
go build -o bin/hello-world main.go

echo "✓ Binary built: bin/hello-world"

# Check if manifest exists
if [ ! -f manifest.yaml ]; then
    echo "Error: manifest.yaml not found"
    exit 1
fi

echo "✓ Package ready for deployment"
echo ""
echo "To create a package, run:"
echo "  cd examples/hello-world"
echo "  tar -czf hello-world-1.0.0.tar.gz manifest.yaml bin/"
