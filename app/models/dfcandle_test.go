package models

import (
	"testing"
	"log"
	// "cryptocurrency/config"
	"time"
)

func TestBackTestEma(t *testing.T) {
	t.Skip()
	df, _ := GetAllCandle("BTC_JPY", time.Minute, 365)
	// logssssss.Printf("%+v", df)
	log.Println(len(df.Candles))
	signalEvents := df.BackTestEma(7, 14)
	log.Printf("%+v", signalEvents)
	caled_profit := 0.0
	for i, p := range signalEvents.Signals {
		log.Println(p.Price)
		if i % 2 == 0 {
			caled_profit = caled_profit - p.Price
		} else {
			caled_profit = caled_profit + p.Price
		}
	}
	log.Println(caled_profit)
	profit := signalEvents.Profit()
	log.Println(profit)

	// perfoamance, bestPeriod1, bestPeriod2 := df.OptimizeEma()
	// log.Println(perfoamance, bestPeriod1, bestPeriod2)
}
