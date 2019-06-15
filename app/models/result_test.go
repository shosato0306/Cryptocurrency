package models

import (
	"testing"
	// "log"
	"time"
)

func TestInsertBuyResult(t *testing.T) {
	t.Skip()
	buyTime, _ := time.Parse("2006-01-02 15:04:05", "2019-06-14 23:15:11")
	buyPrice := 100.00
	coinPriceBuy := 800000.00
	stopLimitPercent := 0.995
	param1 := 3.0
	param2 := 12.0
	var param3 float64
	indicator := "ema"
	exchange := "quoine"
	productCode := "BTC_JPY"
	tradeDuration := "1m"
	dataLimit := 365
	numRanking := 1


	err := InsertBuyResult(buyTime, buyPrice, coinPriceBuy, stopLimitPercent, param1, param2, param3,
		indicator, exchange, productCode, tradeDuration, dataLimit, numRanking)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdeateSellResult(t *testing.T) {
	t.Skip()
	sellTime, _ := time.Parse("2006-01-02 15:04:05", "2019-06-23 02:01:01")
	balance := 1000000.999
	sellPrice := 120.99
	coinPriceSell := 800200.999
	profit := 20.999
	profitPercent := 0.001

	err := UpdateSellResult(sellTime, sellPrice, balance, coinPriceSell, profit, profitPercent)
	if err != nil {
		t.Fatal(err)
	}
}