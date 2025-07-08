#!/bin/bash

# Rebuild entrag with timing features
echo "Rebuilding entrag with timing features..."

# Set up Go environment
export PATH=/usr/local/go/bin:$PATH

# Remove old binary
rm -f entrag

# Build new binary
go build -o entrag ./cmd/entrag/

# Check if build was successful
if [ -f "entrag" ]; then
    echo "✅ Build successful!"
    
    # Copy to global bin directory
    sudo cp entrag /usr/local/bin/entrag
    
    echo "✅ Entrag installed globally!"
    echo "You can now use: entrag ask \"your question\""
else
    echo "❌ Build failed!"
    exit 1
fi 