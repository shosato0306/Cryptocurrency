package main

import (
	"cryptocurrency/app/models"
	"cryptocurrency/config"
	"cryptocurrency/utils"
	"fmt"
)

func main() {
	utils.LoggingSettings(config.Config.LogFile)
	fmt.Println(models.DbConnection)
}
