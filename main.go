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

	text :="Walking down these cobbled streets the morning feels so new clouds are breaking overhead the sunlight’s shining through I’ve carried all my worries but I’ll set them down today the world is wide and waiting I’ll go my own way voices in the corner pub laughter spilling out stories told a thousand times what life is all about the rain may fall tomorrow but tonight the stars will stay with friends around beside me I’ll chase the night away dreams are built on steady hands and songs we choose to sing from London down to Liverpool they echo on the wind no matter where the road bends no matter where it strays my heart is set on moving through brighter lighter days so raise a glass to journeys to lessons still to find to love that keeps on burning to peace within the mind the past is just a shadow tomorrow’s yet to play but here I stand still singing and I’ll sing on my way"
	//voice := tts.Voice("en_us_001")
	outputFile := "output.mp3"

	err := tts.TTS(text, tts.Voice(models.UK_MALE_1), outputFile, false)
	if err != nil {
		log.Fatalf("TTS error: %v", err)
	}

	lg.Println("TTS completed, saved to", outputFile)
}