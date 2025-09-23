#!/usr/bin/env bash

set -e

sudo apt install -y build-essential cmake

ROOT_DIR="$(pwd)"
SCRIPTS_DIR="$ROOT_DIR/scripts"
WHISPER_DIR="$SCRIPTS_DIR/whisper.cpp"
BINDINGS_DIR="$WHISPER_DIR/bindings/go"

echo ">>> Setting up whisper..."

# Clone whisper.cpp if missing
if [ ! -d "$WHISPER_DIR" ]; then
    echo "[*] Cloning whisper.cpp..."
    git clone https://github.com/ggerganov/whisper.cpp.git "$WHISPER_DIR"
else
    echo "[*] whisper.cpp already exists"
fi

# Build whisper library via the Go bindings Makefile
echo "[*] Building whisper via Go bindings..."
cd "$BINDINGS_DIR"
make whisper

# Setup include and library paths
INCLUDE_PATH="$(realpath ../../include):$(realpath ../../ggml/include)"
LIBRARY_PATH="$(realpath ../../build_go/src):$(realpath ../../build_go/ggml/src)"

export C_INCLUDE_PATH="$INCLUDE_PATH"
export LIBRARY_PATH="$LIBRARY_PATH"

echo "C_INCLUDE_PATH=$C_INCLUDE_PATH"
echo "LIBRARY_PATH=$LIBRARY_PATH"

echo ">>> Done setting up whisper ✅ ✅ ✅"

# Go back to project root and run main.go
cd "$ROOT_DIR"
echo "Running main.go..."
go run main.go
