package main

import (
	"cryptocurrency/app/controllers"
	"cryptocurrency/config"
	"cryptocurrency/utils"
)

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	controllers.StreamIngestionData()
	controllers.StartWebServer()
}
