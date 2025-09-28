package main

import (
	"log"
	"sts/internal/models"
	"sts/services/stt"
	"sts/services/tts"
	"sts/utils"
)

func main() {

	utils.Init()

	lg := utils.Logger

	lg.Println("hello sts...")

	// download setup
	utils.SetupFFmpeg()
	utils.SetupModel()

	lg.Println("processing videos")

	stt.ProcessAllVideos(lg)

	text :="hello hi             bonjour "
	//voice := tts.Voice("en_us_001")
	outputFile := "output.mp3"

	err := tts.TTS(text, tts.Voice(models.UK_MALE_1), outputFile, false, lg)
	if err != nil {
		log.Fatalf("TTS error: %v", err)
	}

	lg.Println("TTS completed, saved to", outputFile)
}