package controllers

import (
	"cryptocurrency/app/models"
	"cryptocurrency/bitflyer"
	"cryptocurrency/config"
	"cryptocurrency/quoine"
	"cryptocurrency/slack"
	"log"
	"time"
)

func StreamIngestionData() {
	c := config.Config
	ai := NewAI(c.ProductCode, c.TradeDuration, c.DataLimit, c.UsePercent, c.StopLimitPercent, c.BackTest)
	if c.Exchange == "bitflyer" {
		var tickerChannel = make(chan models.Ticker)
		apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
		go apiClient.GetRealTimeTicker(config.Config.ProductCode, tickerChannel)
		go func() {
			for ticker := range tickerChannel {
				// log.Printf("action=StreamIngestionData, %v", ticker)
				for _, duration := range config.Config.Durations {
					isCreated := models.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
					// 新規に Candle 情報が作成され、なおかつ設定したトレード期間に一致した場合は、
					// インディケータのパラメータの最適化と売買判断を実行する。
					if isCreated == true && duration == config.Config.TradeDuration {
						ai.Trade()
					}
				}
			}
		}()
	} else if c.Exchange == "quoine" {
		var tickerChannel = make(chan *models.Ticker)
		apiClient := quoine.New(config.Config.ApiKey, config.Config.ApiSecret)
		go apiClient.GetRealTimeProduct(config.Config.ProductCode, tickerChannel)
		go func() {
			for ticker := range tickerChannel {
				for _, duration := range config.Config.Durations {
					isCreated := models.CreateCandleWithDuration(*ticker, ticker.ProductCode, duration)
					if isCreated == true && duration == config.Config.TradeDuration {
						// log.Println("### Trade() is called")
						ai.Trade()
					}
				}
			}
		}()
	}
}

func CleanUpRecord() {
	c := config.Config
	go func() {
		for {
			for _, duration := range c.Durations {
				err := models.CleanCandleRecord(c.ProductCode, duration, c.DataLimit)
				if err != nil {
					slack.Notice("notification", "CleanUpRecord failed: " + err.Error())
					log.Fatal(err)
				}
			}
			log.Println("Deletion old records is complete. Wait 30 minutes. ...")
			time.Sleep(time.Minute * 30)
		}
	}()
}
