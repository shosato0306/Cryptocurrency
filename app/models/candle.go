package models

import (
	"cryptocurrency/bitflyer"
	"fmt"
	"log"
	"time"
)

type Candle struct {
	ProductCode string        `json:"product_code" gorm:"-"`
	Duration    time.Duration `json:"duration" gorm:"-"`
	Time        time.Time     `json:"time" gorm:"primary_key"`
	Open        float64       `json:"open"`
	Close       float64       `json:"close"`
	High        float64       `json:"high"`
	Low         float64       `json:"low"`
	Volume      float64       `json:"volume"`
}

func NewCandle(productCode string, duration time.Duration, timeDate time.Time, open, close, high, low, volume float64) *Candle {
	return &Candle{
		productCode,
		duration,
		timeDate,
		open,
		close,
		high,
		low,
		volume,
	}
}

func (c *Candle) GetTableName() string {
	return GetCandleTableName(c.ProductCode, c.Duration)
}

func (c *Candle) Create() error {
	cmd := fmt.Sprintf("INSERT INTO %s (time, open, close, high, low, volume) VALUES (?, ?, ?, ?, ?, ?)", c.GetTableName())
	_, err := DbConnection.Exec(cmd, c.Time.Format(time.RFC3339), c.Open, c.Close, c.High, c.Low, c.Volume)
	if err != nil {
		return err
	}

	// for MySQL
	cmd = fmt.Sprintf("INSERT INTO %s (time, open, close, high, low, volume) VALUES (?, ?, ?, ?, ?, ?);", c.GetTableName())
	_, err = DB.Exec(cmd, c.Time, c.Open, c.Close, c.High, c.Low, c.Volume)
	if err != nil {
		log.Println("Create candle record failed: ", err)
		return err
	}

	return err
}

func (c *Candle) Save() error {
	cmd := fmt.Sprintf("UPDATE %s SET open = ?, close = ?, high = ?, low = ?, volume = ? WHERE time = ?", c.GetTableName())
	_, err := DbConnection.Exec(cmd, c.Open, c.Close, c.High, c.Low, c.Volume, c.Time.Format(time.RFC3339))
	if err != nil {
		return err
	}

	// for MySQL
	cmd = fmt.Sprintf("UPDATE %s SET open = ?, close = ?, high = ?, low = ?, volume = ? WHERE time = ?;", c.GetTableName())
	_, err = DB.Exec(cmd, c.Open, c.Close, c.High, c.Low, c.Volume, c.Time)
	if err != nil {
		log.Println("Update candel record failed: ", err)
		return err
	}

	return err
}

func GetCandle(productCode string, duration time.Duration, dateTime time.Time) *Candle {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf("SELECT time, open, close, high, low, volume FROM  %s WHERE time = ?", tableName)
	row := DbConnection.QueryRow(cmd, dateTime.Format(time.RFC3339))
	var candle Candle
	err := row.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
	if err != nil {
		return nil
	}

	// for MySQL
	cmd = fmt.Sprintf("SELECT time, open, close, high, low, volume FROM  %s WHERE time = ?;", tableName)
	testRow := DB.QueryRow(cmd, dateTime)
	var testCandle Candle
	err = testRow.Scan(&testCandle.Time, &testCandle.Open, &testCandle.Close, &testCandle.High, &testCandle.Low, &testCandle.Volume)
	if err != nil {
		log.Println("Get candle record failed: ", err)
		return nil
	}

	return NewCandle(productCode, duration, candle.Time, candle.Open, candle.Close, candle.High, candle.Low, candle.Volume)
}

func CreateCandleWithDuration(ticker bitflyer.Ticker, productCode string, duration time.Duration) bool {
	currentCandle := GetCandle(productCode, duration, ticker.TruncateDateTime(duration))
	price := ticker.GetMidPrice()
	// DB にまだ対象 duration の Candle 情報が格納されていない場合は、新規に Candle を作成し DB に格納する。
	if currentCandle == nil {
		candle := NewCandle(productCode, duration, ticker.TruncateDateTime(duration),
			price, price, price, price, ticker.Volume)
		err := candle.Create()
		if err != nil {
			log.Println("Record Insert Error: ", err)
		}
		return true
	}

	// DB にすでに対象 duration の Candle 情報が格納されており、最高値、最安値を更新している場合は DB のレコードを更新する。
	if currentCandle.High <= price {
		currentCandle.High = price
	} else if currentCandle.Low >= price {
		currentCandle.Low = price
	}
	currentCandle.Volume += ticker.Volume
	currentCandle.Close = price
	currentCandle.Save()
	return false
}

func GetAllCandle(productCode string, duration time.Duration, limit int) (dfCandle *DataFrameCandle, err error) {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf(`SELECT * FROM (
		SELECT time, open, close, high, low, volume FROM %s ORDER BY time DESC LIMIT ?
		) ORDER BY time ASC;`, tableName)
	rows, err := DbConnection.Query(cmd, limit)
	if err != nil {
		return
	}
	defer rows.Close()

	dfCandle = &DataFrameCandle{}
	dfCandle.ProductCode = productCode
	dfCandle.Duration = duration
	for rows.Next() {
		var candle Candle
		candle.ProductCode = productCode
		candle.Duration = duration
		rows.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
		dfCandle.Candles = append(dfCandle.Candles, candle)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// For MySQL
	cmd = fmt.Sprintf(`SELECT * FROM %s  WHERE time IN (
		SELECT tmp.time FROM (SELECT time FROM %s ORDER BY time DESC LIMIT ?
		) AS tmp) ORDER BY time ASC;`, tableName, tableName)
	rows, err = DB.Query(cmd, limit)
	if err != nil {
		return
	}

	return dfCandle, nil
}

func CleanCandleRecord(productCode string, duration time.Duration, limit int) error {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf("DELETE FROM %s WHERE time NOT IN (SELECT time FROM %s ORDER BY time DESC limit ?)", tableName, tableName)
	_, err := DbConnection.Exec(cmd, limit)
	if err != nil {
		return err
	}

	// for MySQL
	cmd = fmt.Sprintf("DELETE FROM %s WHERE time NOT IN (SELECT tmp.time FROM (SELECT time FROM %s ORDER BY time DESC limit ?) AS tmp);", tableName, tableName)
	_, err = DB.Exec(cmd, limit)
	if err != nil {
		log.Println("Delete records failed: ", err)
		return err
	}

	return err
}
