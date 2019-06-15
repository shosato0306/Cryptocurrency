package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"cryptocurrency/config"
)

type SignalEvent struct {
	Time        time.Time `json:"time" gorm:"primary_key"`
	ProductCode string    `json:"product_code"`
	Side        string    `json:"side"`
	Price       float64   `json:"price"`
	Size        float64   `json:"size"`
}

func (s *SignalEvent) Save() bool {
	cmd := fmt.Sprintf("INSERT INTO %s (time, product_code, side, price, size) VALUES (?, ?, ?, ?, ?);", tableNameSignalEvents)
	_, err := DB.Exec(cmd, s.Time, s.ProductCode, s.Side, s.Price, s.Size)
	if err != nil {
		// time が主キーではなくなるため以下のチェックは不要になる予定
		if strings.Contains(err.Error(), "Duplicate entry") {
			log.Println("Insert record error: ", err)
			return true
		}
		return false
	}

	return true
}

type SignalEvents struct {
	Signals []SignalEvent `json:"signals,omitempty"`
}

// Backtest 時には新しい SignalEvents オブジェクトのアドレスを返す
func NewSignalEvents() *SignalEvents {
	return &SignalEvents{}
}

// 実際に取引を行っている場合(Backtest == False)の場合の処理
// DB から指定した個数分だけ最新の signalevent 情報を取得する
func GetSignalEventsByCount(loadEvents int) *SignalEvents {
	// cmd := fmt.Sprintf(`SELECT * FROM %s WHERE time IN (
    //     SELECT tmp.time FROM (SELECT time FROM %s WHERE product_code = ? ORDER BY time DESC LIMIT ? ) AS tmp)
	// 	ORDER BY time ASC;`, tableNameSignalEvents, tableNameSignalEvents)
	// cmd := fmt.Sprintf(`SELECT * FROM %s WHERE id IN (
	cmd := fmt.Sprintf(`SELECT time, product_code, side, price, size FROM %s WHERE id IN (
        SELECT tmp.id FROM (SELECT id FROM %s WHERE product_code = ? ORDER BY id DESC LIMIT ? ) AS tmp)
		ORDER BY id ASC;`, tableNameSignalEvents, tableNameSignalEvents)

	rows, err := DB.Query(cmd, config.Config.ProductCode, loadEvents)
	if err != nil {
		log.Println("Get signalevent error: ", err)
		return nil
	}
	defer rows.Close()

	var signalEvents SignalEvents
	for rows.Next() {
		var signalEvent SignalEvent
		rows.Scan(&signalEvent.Time, &signalEvent.ProductCode, &signalEvent.Side, &signalEvent.Price, &signalEvent.Size)
		signalEvents.Signals = append(signalEvents.Signals, signalEvent)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil
	}
	return &signalEvents
}

func GetSignalEventsAfterTime(timeTime time.Time) *SignalEvents {
	// cmd := fmt.Sprintf(`SELECT * FROM %s WHERE time IN (
	cmd := fmt.Sprintf(`SELECT time, product_code, side, price, size FROM %s WHERE time IN (
			SELECT time FROM %s
			WHERE time >= ?
			ORDER BY time DESC
		) ORDER BY time ASC;`, tableNameSignalEvents, tableNameSignalEvents)

	rows, err := DB.Query(cmd, timeTime.Format(time.RFC3339))
	if err != nil {
		fmt.Println("GetSignalEventsAfterTime failed: ", err)
		return nil
	}
	defer rows.Close()

	var signalEvents SignalEvents
	for rows.Next() {
		var signalEvent SignalEvent
		rows.Scan(&signalEvent.Time, &signalEvent.ProductCode, &signalEvent.Side, &signalEvent.Price, &signalEvent.Size)
		signalEvents.Signals = append(signalEvents.Signals, signalEvent)
	}
	return &signalEvents
}

func (s *SignalEvents) CanBuy(time time.Time) bool {
	lenSignals := len(s.Signals)
	if lenSignals == 0 {
		return true
	}

	lastSignal := s.Signals[lenSignals-1]
	// if lastSignal.Side == "SELL" && lastSignal.Time.Before(time) {
	if lastSignal.Side == "SELL" && (lastSignal.Time.Before(time) || lastSignal.Time.Equal(time)) {
		return true
	}
	return false
}

func (s *SignalEvents) CanSell(time time.Time) bool {
	lenSignals := len(s.Signals)
	if lenSignals == 0 {
		return false
	}

	lastSignal := s.Signals[lenSignals-1]
	// if lastSignal.Side == "BUY" && lastSignal.Time.Before(time) {
	if lastSignal.Side == "BUY" && (lastSignal.Time.Before(time) || lastSignal.Time.Equal(time)) {
		return true
	}
	return false
}

func (s *SignalEvents) Buy(ProductCode string, time time.Time, price, size float64, save bool) bool {
	if !s.CanBuy(time) {
		return false
	}
	signalEvent := SignalEvent{
		ProductCode: ProductCode,
		Time:        time,
		Side:        "BUY",
		Price:       price,
		Size:        size,
	}
	if save {
		signalEvent.Save()
	}
	s.Signals = append(s.Signals, signalEvent)
	return true
}

func (s *SignalEvents) Sell(productCode string, time time.Time, price, size float64, save bool) bool {

	if !s.CanSell(time) {
		return false
	}

	signalEvent := SignalEvent{
		ProductCode: productCode,
		Time:        time,
		Side:        "SELL",
		Price:       price,
		Size:        size,
	}
	if save {
		signalEvent.Save()
	}
	s.Signals = append(s.Signals, signalEvent)
	return true
}

func (s *SignalEvents) Profit() float64 {
	total := 0.0
	beforeSell := 0.0
	isHolding := false
	for i, signalEvent := range s.Signals {
		if i == 0 && signalEvent.Side == "SELL" {
			continue
		}
		if signalEvent.Side == "BUY" {
			total -= signalEvent.Price * signalEvent.Size
			isHolding = true
		}
		if signalEvent.Side == "SELL" {
			total += signalEvent.Price * signalEvent.Size
			isHolding = false
			beforeSell = total
		}
	}
	if isHolding == true {
		return beforeSell
	}
	return total
}

func (s SignalEvents) MarshalJSON() ([]byte, error) {
	value, err := json.Marshal(&struct {
		Signals []SignalEvent `json:"signals,omitempty"`
		Profit  float64       `json:"profit,omitempty"`
	}{
		Signals: s.Signals,
		Profit:  s.Profit(),
	})
	if err != nil {
		return nil, err
	}
	return value, err
}

func (s *SignalEvents) CollectAfter(time time.Time) *SignalEvents {
	for i, signal := range s.Signals {
		if time.After(signal.Time) {
			continue
		}
		return &SignalEvents{Signals: s.Signals[i:]}
	}
	return nil
}
