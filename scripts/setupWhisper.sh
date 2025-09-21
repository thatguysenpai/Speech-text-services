#!/usr/bin/env bash
set -e

echo "[*] Starting setup..."

if [ ! -d "whisper.cpp" ]; then
    echo "[*] Cloning whisper.cpp..."
    git clone https://github.com/ggerganov/whisper.cpp.git
else
    echo "[*] whisper.cpp already exists"
fi


echo "[*] Building whisper.cpp..."
cd whisper.cpp
make
cd ..

WHISPER_DIR="$(pwd)/whisper.cpp"
export CGO_CFLAGS="-I$WHISPER_DIR"
export CGO_LDFLAGS="-L$WHISPER_DIR -lwhisper"

echo "[*] CGO_CFLAGS set to: $CGO_CFLAGS"
echo "[*] CGO_LDFLAGS set to: $CGO_LDFLAGS"


echo "[*] Setup completed ✅"
