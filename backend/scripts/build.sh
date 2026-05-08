#!/bin/bash
set -e

echo "Building MCSPhotoIndex..."
echo ""
echo "Step 1: Building Frontend (npm)..."
cd ../frontend
npm run build
echo "Frontend build complete!"
cd ../backend
echo ""
echo "Step 2: Generating TLS certificate..."
go run ./cmd/gencert
echo "TLS certificate generation complete!"
echo ""
echo "Step 3: Building Go binaries..."
go build -ldflags="-s -w" -o ./bin/schematic2 ./cmd/server
echo ""
echo "Build complete!"
