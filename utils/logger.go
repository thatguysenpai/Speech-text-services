package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var Logger *log.Logger

// Init initializes the logger and sets it as global
func Init() {
	// Ensure logs directory exists
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatalf("failed to create log directory: %v", err)
		}
	}

	// Log file (rotated daily)
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	// Log to both file and stdout
	mw := io.MultiWriter(os.Stdout, f)

	Logger = log.New(mw, "", log.Ldate|log.Ltime|log.Lshortfile)
	Logger.Println("Logger initialized ✅")
}
