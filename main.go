package main

import (
	"log"

	"sts/internal/config"
	"sts/internal/models"
	"sts/services/stt"
	"sts/services/tts"
	"sts/utils"
)

func main() {

	// Setup file system Logger
	utils.Init()
	lg := utils.Logger
	lg.Println("hello sts...")

	// Setup Dependencies
	utils.SetupFFmpeg()
	utils.SetupModel()
	utils.SetupFFprobe(lg)

	// Setup env
	config.LoadEnv(lg)

	// Test

	lg.Println("processing videos")
	stt.ProcessAllVideos(lg)

	text := "hello hi bonjour "
	//voice := tts.Voice("en_us_001")
	outputFile := "output.mp3"

	err := tts.TTS(text, tts.Voice(models.UK_MALE_1), outputFile, false, lg)
	if err != nil {
		log.Fatalf("TTS error: %v", err)
	}

	lg.Println("TTS completed, saved to", outputFile)
}
