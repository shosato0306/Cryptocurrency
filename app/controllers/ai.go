package controllers

import (
	"log"
	"math"
	"strings"
	"time"

	"github.com/markcheno/go-talib"

	"golang.org/x/sync/semaphore"

	"cryptocurrency/app/models"
	"cryptocurrency/bitflyer"
	"cryptocurrency/config"
	"cryptocurrency/tradingalgo"
	"cryptocurrency/slack"
)

const (
	ApiFeePercent = 0.0012
)

type AI struct {
	API                  *bitflyer.APIClient
	ProductCode          string
	CurrencyCode         string
	CoinCode             string
	UsePercent           float64
	MinuteToExpires      int
	Duration             time.Duration
	PastPeriod           int
	SignalEvents         *models.SignalEvents
	OptimizedTradeParams *models.TradeParams
	TradeSemaphore       *semaphore.Weighted
	StopLimit            float64
	StopLimitPercent     float64
	BackTest             bool
	StartTrade           time.Time
}

// TODO mutex, singleton
var Ai *AI

func NewAI(productCode string, duration time.Duration, pastPeriod int, UsePercent, stopLimitPercent float64, backTest bool) *AI {
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	var signalEvents *models.SignalEvents
	if backTest {
		signalEvents = models.NewSignalEvents()
	} else {
		// DB に格納されている最新の signalevent 情報を一つ取得する
		signalEvents = models.GetSignalEventsByCount(1)
	}
	codes := strings.Split(productCode, "_")
	Ai = &AI{
		API:              apiClient,
		ProductCode:      productCode,
		CoinCode:         codes[0],
		CurrencyCode:     codes[1],
		UsePercent:       UsePercent,
		MinuteToExpires:  1,
		PastPeriod:       pastPeriod,
		Duration:         duration,
		SignalEvents:     signalEvents,
		TradeSemaphore:   semaphore.NewWeighted(1),
		BackTest:         backTest,
		StartTrade:       time.Now(),
		StopLimitPercent: stopLimitPercent,
	}
	Ai.UpdateOptimizeParams(false)
	return Ai
}

func (ai *AI) UpdateOptimizeParams(isContinue bool) {
	df, _ := models.GetAllCandle(ai.ProductCode, ai.Duration, ai.PastPeriod)
	ai.OptimizedTradeParams = df.OptimizeParams()
	log.Printf("optimized_trade_params=%+v", ai.OptimizedTradeParams)
	if ai.OptimizedTradeParams == nil && isContinue && !ai.BackTest {
		log.Print("status_no_params")
		time.Sleep(5 * ai.Duration)
		ai.UpdateOptimizeParams(isContinue)
	}
}

func (ai *AI) Buy(candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool) {
	if ai.BackTest {
		couldBuy := ai.SignalEvents.Buy(ai.ProductCode, candle.Time, candle.Close, 1.0, false)
		return "", couldBuy
	}

	if ai.StartTrade.After(candle.Time) {
		return
	}

	if !ai.SignalEvents.CanBuy(candle.Time) {
		return
	}

	availableCurrency, _ := ai.GetAvailableBalance()
	useCurrency := availableCurrency * ai.UsePercent
	ticker, err := ai.API.GetTicker(ai.ProductCode)
	if err != nil {
		return
	}
	size := 1 / (ticker.BestAsk / useCurrency)
	size = ai.AdjustSize(size)

	order := &bitflyer.Order{
		ProductCode:     ai.ProductCode,
		ChildOrderType:  "MARKET",
		Side:            "BUY",
		Size:            size,
		MinuteToExpires: ai.MinuteToExpires,
		TimeInForce:     "GTC",
	}
	log.Printf("status=buy candle=%+v order=%+v", candle, order)
	resp, err := ai.API.SendOrder(order)
	if err != nil {
		slack.Notice("notification", "Send order failed: " + err.Error())
		log.Println("Send order failed: ", err)
		return
	}
	if resp.ChildOrderAcceptanceID == "" {
		// Insufficient fund
		slack.Notice("notification", "Insufficient fund")
		log.Printf("order=%+v status=no_id", order)
		return
	}
	childOrderAcceptanceID = resp.ChildOrderAcceptanceID
	isOrderCompleted = ai.WaitUntilOrderComplete(childOrderAcceptanceID, candle.Time)
	return childOrderAcceptanceID, isOrderCompleted
}

func (ai *AI) Sell(candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool) {
	if ai.BackTest {
		couldSell := ai.SignalEvents.Sell(ai.ProductCode, candle.Time, candle.Close, 1.0, false)
		return "", couldSell
	}

	if ai.StartTrade.After(candle.Time) {
		return
	}

	if !ai.SignalEvents.CanSell(candle.Time) {
		return
	}

	_, availableCoin := ai.GetAvailableBalance()
	size := ai.AdjustSize(availableCoin)

	order := &bitflyer.Order{
		ProductCode:     ai.ProductCode,
		ChildOrderType:  "MARKET",
		Side:            "SELL",
		Size:            size,
		MinuteToExpires: ai.MinuteToExpires,
		TimeInForce:     "GTC",
	}
	log.Printf("status=sell candle=%+v order=%+v", candle, order)
	resp, err := ai.API.SendOrder(order)
	if err != nil {
		slack.Notice("notification", "Send order failed: " + err.Error())
		log.Println("Send order failed: ", err)
		return
	}
	if resp.ChildOrderAcceptanceID == "" {
		// Insufficient fund
		slack.Notice("notification", "Insufficient fund")
		log.Printf("order=%+v status=no_id", order)
		return
	}
	childOrderAcceptanceID = resp.ChildOrderAcceptanceID

	isOrderCompleted = ai.WaitUntilOrderComplete(childOrderAcceptanceID, candle.Time)
	return childOrderAcceptanceID, isOrderCompleted
}

func (ai *AI) Trade() {
	isAcquire := ai.TradeSemaphore.TryAcquire(1)
	if !isAcquire {
		slack.Notice("notification", "Could not get trade lock")
		log.Println("Could not get trade lock")
		return
	}
	defer ai.TradeSemaphore.Release(1)
	params := ai.OptimizedTradeParams
	if params == nil {
		return
	}
	df, _ := models.GetAllCandle(ai.ProductCode, ai.Duration, ai.PastPeriod)
	lenCandles := len(df.Candles)

	var emaValues1 []float64
	var emaValues2 []float64
	if params.EmaEnable {
		emaValues1 = talib.Ema(df.Closes(), params.EmaPeriod1)
		emaValues2 = talib.Ema(df.Closes(), params.EmaPeriod2)
	}

	var bbUp []float64
	var bbDown []float64
	if params.BbEnable {
		bbUp, _, bbDown = talib.BBands(df.Closes(), params.BbN, params.BbK, params.BbK, 0)
	}

	var tenkan, kijun, senkouA, senkouB, chikou []float64
	if params.IchimokuEnable {
		tenkan, kijun, senkouA, senkouB, chikou = tradingalgo.IchimokuCloud(df.Closes())
	}

	var outMACD, outMACDSignal []float64
	if params.MacdEnable {
		outMACD, outMACDSignal, _ = talib.Macd(df.Closes(), params.MacdFastPeriod, params.MacdSlowPeriod, params.MacdSignalPeriod)
	}

	var rsiValues []float64
	if params.RsiEnable {
		rsiValues = talib.Rsi(df.Closes(), params.RsiPeriod)
	}

	for i := 1; i < lenCandles; i++ {
		buyPoint, sellPoint := 0, 0
		if params.EmaEnable && params.EmaPeriod1 <= i && params.EmaPeriod2 <= i {
			if emaValues1[i-1] < emaValues2[i-1] && emaValues1[i] >= emaValues2[i] {
				buyPoint++
			}

			if emaValues1[i-1] > emaValues2[i-1] && emaValues1[i] <= emaValues2[i] {
				sellPoint++
			}
		}

		if params.BbEnable && params.BbN <= i {
			if bbDown[i-1] > df.Candles[i-1].Close && bbDown[i] <= df.Candles[i].Close {
				buyPoint++
			}

			if bbUp[i-1] < df.Candles[i-1].Close && bbUp[i] >= df.Candles[i].Close {
				sellPoint++
			}
		}

		if params.MacdEnable {
			if outMACD[i] < 0 && outMACDSignal[i] < 0 && outMACD[i-1] < outMACDSignal[i-1] && outMACD[i] >= outMACDSignal[i] {
				buyPoint++
			}

			if outMACD[i] > 0 && outMACDSignal[i] > 0 && outMACD[i-1] > outMACDSignal[i-1] && outMACD[i] <= outMACDSignal[i] {
				sellPoint++
			}
		}

		if params.IchimokuEnable {
			if chikou[i-1] < df.Candles[i-1].High && chikou[i] >= df.Candles[i].High &&
				senkouA[i] < df.Candles[i].Low && senkouB[i] < df.Candles[i].Low &&
				tenkan[i] > kijun[i] {
				buyPoint++
			}

			if chikou[i-1] > df.Candles[i-1].Low && chikou[i] <= df.Candles[i].Low &&
				senkouA[i] > df.Candles[i].High && senkouB[i] > df.Candles[i].High &&
				tenkan[i] < kijun[i] {
				sellPoint++
			}
		}

		if params.RsiEnable && rsiValues[i-1] != 0 && rsiValues[i-1] != 100 {
			if rsiValues[i-1] < params.RsiBuyThread && rsiValues[i] >= params.RsiBuyThread {
				buyPoint++
			}

			if rsiValues[i-1] > params.RsiSellThread && rsiValues[i] <= params.RsiSellThread {
				sellPoint++
			}
		}

		if buyPoint > 0 {
			_, isOrderCompleted := ai.Buy(df.Candles[i])
			if !isOrderCompleted {
				continue
			}
			ai.StopLimit = df.Candles[i].Close * ai.StopLimitPercent
		}

		if sellPoint > 0 || ai.StopLimit > df.Candles[i].Close {
			_, isOrderCompleted := ai.Sell(df.Candles[i])
			if !isOrderCompleted {
				continue
			}
			ai.StopLimit = 0.0
			ai.UpdateOptimizeParams(true)
		}
	}
}

func (ai *AI) GetAvailableBalance() (availableCurrency, availableCoin float64) {
	balances, err := ai.API.GetBalance()
	if err != nil {
		return
	}
	for _, balance := range balances {
		if balance.CurrentCode == ai.CurrencyCode {
			availableCurrency = balance.Available
		} else if balance.CurrentCode == ai.CoinCode {
			availableCoin = balance.Available
		}
	}
	return availableCurrency, availableCoin
}

func (ai *AI) AdjustSize(size float64) float64 {
	fee := size * ApiFeePercent
	size = size - fee
	return math.Floor(size*10000) / 10000
}

func (ai *AI) WaitUntilOrderComplete(childOrderAcceptanceID string, executeTime time.Time) bool {
	params := map[string]string{
		"product_code":              ai.ProductCode,
		"child_order_acceptance_id": childOrderAcceptanceID,
	}
	expire := time.After(time.Minute + (20 * time.Second))
	interval := time.Tick(15 * time.Second)
	return func() bool {
		for {
			select {
			case <-expire:
				return false
			case <-interval:
				listOrders, err := ai.API.ListOrder(params)
				if err != nil {
					return false
				}
				if len(listOrders) == 0 {
					return false
				}
				order := listOrders[0]
				if order.ChildOrderState == "COMPLETED" {
					if order.Side == "BUY" {
						couldBuy := ai.SignalEvents.Buy(ai.ProductCode, executeTime, order.AveragePrice, order.Size, true)
						if !couldBuy {
							slack.Notice("trade", "BUY process completed !")
							log.Printf("status=buy childOrderAcceptanceID=%s order=%+v", childOrderAcceptanceID, order)
						}
						return couldBuy
					}
					if order.Side == "SELL" {
						couldSell := ai.SignalEvents.Sell(ai.ProductCode, executeTime, order.AveragePrice, order.Size, true)
						if !couldSell {
							slack.Notice("trade", "SELL process completed !")
							log.Printf("status=sell childOrderAcceptanceID=%s order=%+v", childOrderAcceptanceID, order)
						}
						return couldSell
					}
					return false
				}
			}
		}
	}()
}
