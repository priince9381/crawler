#!/bin/bash
MAIN_FILE="main.go"

# Define the output binary name
OUTPUT_BINARY="app"

# Check if Go is installed
if ! command -v go &> /dev/null
then
    echo "Go is not installed. Please install Go and try again."
    exit 1
fi

# Tidy up dependencies
echo "Tidying up Go modules..."
go mod tidy

# Build the Go program
echo "Building Go program..."
go build -o $OUTPUT_BINARY $MAIN_FILE

# Check if build was successful
if [ ! -f "$OUTPUT_BINARY" ]; then
    echo "Build failed."
    exit 1
fi

# Run the binary
echo "Running the Go application..."
./$OUTPUT_BINARY
