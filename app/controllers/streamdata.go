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
						// ai.Trade()
						
					}
				}
			}
		}()
	} else if c.Exchange == "quoine" {
		var tickerChannel = make(chan *models.Ticker)
		apiClient := quoine.New(config.Config.ApiKey, config.Config.ApiSecret)
		var counter int
		bought_in_same_candle :=  false
		sold_in_same_candle := false
		is_holding := false
		go apiClient.GetRealTimeProduct(config.Config.ProductCode, tickerChannel)
		// is_ordered := false
		// call_count := 0
		go func() {
			for ticker := range tickerChannel {
				// log.Println(counter)
				counter += 1 
				// if counter >= 180 {
				for _, duration := range config.Config.Durations {
					// isCreated := models.CreateCandleWithDuration(*ticker, ticker.ProductCode, duration)

					isCreated := models.CreateCandleWithDuration(*ticker, ticker.ProductCode, duration)
					if isCreated == true && duration == config.Config.TradeDuration {
						bought_in_same_candle = false
						sold_in_same_candle = false
					}
					if duration == config.Config.TradeDuration {
						// log.Println("### Trade() is called")
						// is_during_buy := false
						// is_ordered = ai.Trade()
						if is_holding && counter >= 3 || counter >= 40 {
							bought_in_same_candle, sold_in_same_candle, is_holding = ai.Trade(bought_in_same_candle, sold_in_same_candle, is_holding)
							counter = 0
						} 
						// log.Println("ai.Trade()...")
						// if call_count >= 5 && is_during_buy != false {
						// 	ai.UpdateOptimizeParams(true)
						// }
					}			
				}
				// counter = 0
				// }
			}
		}()
	}
}

func CleanUpRecord() {
	c := config.Config
	go func() {
		for {
			for _, duration := range c.Durations {
				err := models.CleanCandleRecord(c.ProductCode, duration, 2000)
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
