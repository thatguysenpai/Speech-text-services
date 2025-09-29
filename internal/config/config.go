package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	GEMINI_API_KEY = "AIzaSyCtrp1MAZHVbIexNel_omP5aMuJeByQfn4"
)

func LoadEnv(log *log.Logger) {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] No .env file found, relying on system environment")
	}
	GEMINI_API_KEY = os.Getenv("GEMINI_API_KEY")
	if GEMINI_API_KEY == "" {
		log.Fatal("[ERROR] GEMINI_API_KEY not set in environment")
	}


	log.Println("ENV loaded successfully")
}
