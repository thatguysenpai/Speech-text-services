package main

import (
	"sts/services/stt"
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
}
