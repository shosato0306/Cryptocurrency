package main

import (
	"cryptocurrency/app/models"
	"fmt"
	"time"
)

// func main() {
// 	utils.LoggingSettings(config.Config.LogFile)
// 	controllers.StreamIngestionData()
// 	controllers.StartWebServer()
// }

func main() {
	s := models.NewSignalEvents()
	df, _ := models.GetAllCandle("BTC_JPY", time.Minute, 10)
	c1 := df.Candles[0]
	c2 := df.Candles[5]
	s.Buy("BTC_JPY", c1.Time.UTC(), c1.Close, 1.0, true)
	s.Sell("BTC_JPY", c2.Time.UTC(), c2.Close, 1.0, true)
	fmt.Println(models.GetSignalEventsByCount(1))
	fmt.Println(models.GetSignalEventsAfterTime(c1.Time))
	fmt.Println(s.CollectAfter(time.Now().UTC()))
	fmt.Println(s.CollectAfter(c1.Time))
}
