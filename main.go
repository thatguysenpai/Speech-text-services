package main

import (
	"sts/utils"
)

func main() {

	utils.Init()

	lg := utils.Logger

	lg.Println("hello sts...")

	utils.SetupFFmpeg()
}
