package controllers

import (
	"cryptocurrency/app/models"
	"cryptocurrency/bitflyer"
	"cryptocurrency/config"
)

func StreamIngestionData() {
	c := config.Config
	ai := NewAI(c.ProductCode, c.TradeDuration, c.DataLimit, c.UsePercent, c.StopLimitPercent, c.BackTest)

	var tickerChannl = make(chan bitflyer.Ticker)
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	go apiClient.GetRealTimeTicker(config.Config.ProductCode, tickerChannl)
	go func() {
		for ticker := range tickerChannl {
			// log.Printf("action=StreamIngestionData, %v", ticker)
			for _, duration := range config.Config.Durations {
				isCreated := models.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
				// 新規に Candle 情報が作成され、なおかつ設定したトレード期間に一致した場合は、
				// トレードを実行する。
				if isCreated == true && duration == config.Config.TradeDuration {
					ai.Trade()
				}
			}
		}
	}()
}
