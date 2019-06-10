package models

import (
	"log"
	"time"
)

type Ticker struct {
	ProductCode string `json:"product_code"`
	Timestamp   string `json:"timestamp"`
	TickID      int    `json:"tick_id"`
	// Bid は買値
	BestBid float64 `json:"best_bid"`
	// Ask は売値
	BestAsk         float64 `json:"best_ask"`
	BestBidSize     float64 `json:"best_bid_size"`
	BestAskSize     float64 `json:"best_ask_size"`
	TotalBidDepth   float64 `json:"total_bid_depth"`
	TotalAskDepth   float64 `json:"total_ask_depth"`
	Ltp             float64 `json:"ltp"`
	Volume          float64 `json:"volume"`
	VolumeByProduct float64 `json:"volume_by_product"`
}

func NewTicker(productCode, timeStamp string, bestBid, bestAsk, volume float64) *Ticker {
	Ticker := &Ticker{
		ProductCode:     productCode,
		Timestamp:       timeStamp,
		TickID:          0,
		BestBid:         bestBid,
		BestAsk:         bestAsk,
		BestBidSize:     0,
		BestAskSize:     0,
		TotalBidDepth:   0,
		TotalAskDepth:   0,
		Ltp:             0,
		Volume:          volume,
		VolumeByProduct: 0,
	}
	return Ticker
}

// 売値と買値の中間の値を取得
func (t *Ticker) GetMidPrice() float64 {
	return (t.BestBid + t.BestAsk) / 2
}

func (t *Ticker) DateTime() time.Time {
	dateTime, err := time.Parse(time.RFC3339, t.Timestamp)
	if err != nil {
		log.Printf("action=DateTime, err=%s", err.Error())
	}
	return dateTime
}

func (t *Ticker) TruncateDateTime(duration time.Duration) time.Time {
	return t.DateTime().Truncate(duration)
}