package main

import (
	"log"
	"trading/config"
	"trading/utils"
)

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	log.Println("test")
}
