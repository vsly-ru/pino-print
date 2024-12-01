#!/bin/bash

# Exit on any error
set -e

APP_NAME="pino-print"
INSTALL_DIR="/usr/local/bin"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

GOOS=$(uname -s | tr '[:upper:]' '[:lower:]')
GOARCH=$(uname -m)

case $GOARCH in
    x86_64)
        GOARCH=amd64
        ;;
    aarch64)
        GOARCH=arm64
        ;;
esac

echo "Building $APP_NAME for $GOOS/$GOARCH"
GOOS=$GOOS GOARCH=$GOARCH go build -o $APP_NAME main.go

echo "Installing $APP_NAME to $INSTALL_DIR"

# Check if we have write permissions to INSTALL_DIR
if [ -w "$INSTALL_DIR" ]; then
    mv $APP_NAME "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$APP_NAME"
else
    # If we don't, use sudo
    echo "Requesting sudo privileges to install to $INSTALL_DIR"
    sudo mv $APP_NAME "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$APP_NAME"
fi


# Verify installation
if command -v $APP_NAME &> /dev/null; then
    echo "✅ Installation complete! You can now use '$APP_NAME' command from anywhere in your terminal."
    $APP_NAME -v
else
    echo "⚠️  Installation succeeded but $APP_NAME is not in PATH"
    echo "You might need to add $INSTALL_DIR to your PATH"
fi