package models

import (
	"time"

	"github.com/markcheno/go-talib"
)

type DataFrameCandle struct {
	ProductCode string        `json:"product_code"`
	Duration    time.Duration `json:"duration"`
	Candles     []Candle      `json:"candles"`
	Smas        []Sma         `json:"smas,omitempty"`
	Emas        []Ema         `json:"emas,omitempty"`
	BBands      *BBands       `json:"bbands,omitempty"`
}

type Sma struct {
	Period int       `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}

type Ema struct {
	Period int       `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}

type BBands struct {
	N    int       `json:"n,omitempty"`
	K    float64   `json:"k,omitempty"`
	Up   []float64 `json:"up,omitempty"`
	Mid  []float64 `json:"mid,omitempty"`
	Down []float64 `json:"down,omitempty"`
}

func (df *DataFrameCandle) Times() []time.Time {
	s := make([]time.Time, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Time
	}
	return s
}

func (df *DataFrameCandle) Opens() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Open
	}
	return s
}

func (df *DataFrameCandle) Closes() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Close
	}
	return s
}

func (df *DataFrameCandle) Highs() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.High
	}
	return s
}

func (df *DataFrameCandle) Low() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Low
	}
	return s
}

func (df *DataFrameCandle) Volume() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Volume
	}
	return s
}

func (df *DataFrameCandle) AddSma(period int) bool {
	if len(df.Candles) > period {
		df.Smas = append(df.Smas, Sma{
			Period: period,
			Values: talib.Sma(df.Closes(), period),
		})
		return true
	}
	return false
}

func (df *DataFrameCandle) AddEma(period int) bool {
	if len(df.Candles) > period {
		df.Emas = append(df.Emas, Ema{
			Period: period,
			Values: talib.Ema(df.Closes(), period),
		})
		return true
	}
	return false
}

func (df *DataFrameCandle) AddBBands(n int, k float64) bool {
	if n <= len(df.Closes()) {
		up, mid, down := talib.BBands(df.Closes(), n, k, k, 0)
		df.BBands = &BBands{
			N:    n,
			K:    k,
			Up:   up,
			Mid:  mid,
			Down: down,
		}
		return true
	}
	return false
}
