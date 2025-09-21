#!/usr/bin/env bash
CGO_CFLAGS="-I$(pwd)/whisper.cpp" CGO_LDFLAGS="-L$(pwd)/whisper.cpp -lwhisper" go run main.go
