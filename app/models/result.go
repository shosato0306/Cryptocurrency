package models

import (
	"fmt"
	"time"
	"log"
)

func InsertBuyResult(buyTime time.Time, buyPrice, coinPriceBuy, stopLimitPercent, param1, param2, param3 float64,
	indicator, exchange, productCode, tradeDuration, refDuration1, refDuration2 string, dataLimit, numRanking int) error {
	var count int
	err := DB.QueryRow("SELECT count(*) FROM results WHERE sell_price IS NULL;").Scan(&count)
	if err != nil {
		log.Println(err)
		return err
	}
	if count >= 1 {
		log.Println("InsertBuyResult is failed: Buy record is already exist.")
		return nil
	}
	
	cmd := fmt.Sprintln(`
	INSERT INTO results (buy_time, buy_price, coin_price_buy, 
		product_code, exchange, stop_limit_percent, indicator,
		param1, param2, param3,
		data_limit, trade_duration, ref_duration1, ref_duration2, num_ranking)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	_, err = DB.Exec(cmd, buyTime, buyPrice, coinPriceBuy, productCode, exchange, stopLimitPercent,
				indicator, param1, param2, param3, dataLimit, tradeDuration, refDuration1, refDuration2, numRanking)
	if err != nil {
		log.Println("Insert result record error: ", err.Error())
		return err
	}
	return nil
}

func UpdateSellResult(sellTime time.Time, sellPrice, balance, coinPriceSell float64) error {
	cmd := fmt.Sprintln(`UPDATE results SET sell_time=?, sell_price=?,
		balance=?, coin_price_sell=? WHERE sell_price IS NULL;`)
	_, err := DB.Exec(cmd, sellTime, sellPrice, balance, coinPriceSell)
	if err != nil {
		log.Println("Update result record error: ", err.Error())
		return err
	}
	return nil
}


