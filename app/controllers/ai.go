package controllers

import (
	"strconv"
	"cryptocurrency/quoine"
	"log"
	"math"
	"strings"
	"time"
	"os"

	"github.com/markcheno/go-talib"

	"golang.org/x/sync/semaphore"

	"cryptocurrency/app/models"
	"cryptocurrency/bitflyer"
	"cryptocurrency/config"
	"cryptocurrency/tradingalgo"
	"cryptocurrency/slack"
)

// const (
// 	ApiFeePercent = 0.0012
// )

type API interface {
	GetTicker(string) (*models.Ticker, error)
	GetBalance() ([]models.Balance, error)
	SendOrder(*models.Order) (string, error)
	ListOrder(map[string]string) ([]models.Order, error) 
}

type AI struct {
	API                  API
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
	var apiClient API
	if config.Config.Exchange == "bitflyer" {
		apiClient = bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	} else if config.Config.Exchange == "quoine" {
		apiClient = quoine.New(config.Config.ApiKey, config.Config.ApiSecret)
	}

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
	log.Println("### UpdateOptimizedParams() is called....")
	log.Printf("optimized_trade_params=%+v", ai.OptimizedTradeParams)
	if ai.OptimizedTradeParams == nil && isContinue && !ai.BackTest {
		log.Print("status_no_params")
		time.Sleep(5 * ai.Duration)
		ai.UpdateOptimizeParams(isContinue)
	}
}

func (ai *AI) GetIndicatorInfo() (indicator string, param1, param2, param3 float64) {
	if ai.OptimizedTradeParams.EmaEnable {
		indicator = "Ema"
		param1 = float64(ai.OptimizedTradeParams.EmaPeriod1)
		param2 = float64(ai.OptimizedTradeParams.EmaPeriod2)
	} else if ai.OptimizedTradeParams.BbEnable {
		indicator = "Bband"
		param1 = float64(ai.OptimizedTradeParams.BbN)
		param2 = float64(ai.OptimizedTradeParams.BbK)
	} else if ai.OptimizedTradeParams.MacdEnable {
		indicator = "Macd"
		param1 = float64(ai.OptimizedTradeParams.MacdFastPeriod)
		param2 = float64(ai.OptimizedTradeParams.MacdSlowPeriod)
		param3 = float64(ai.OptimizedTradeParams.MacdSignalPeriod)
	} else if ai.OptimizedTradeParams.RsiEnable {
		indicator = "Rsi"
		param1 = float64(ai.OptimizedTradeParams.RsiPeriod)
		param2 = float64(ai.OptimizedTradeParams.RsiBuyThread)
		param3 = float64(ai.OptimizedTradeParams.RsiSellThread)
	}
	return indicator, param1, param2, param3
}

func (ai *AI) Buy(candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool) {
	if ai.BackTest {
		couldBuy := ai.SignalEvents.Buy(ai.ProductCode, candle.Time, candle.Close, 1.0, false)
		return "", couldBuy
	}

	if ai.StartTrade.After(candle.Time) {
		// log.Println("ai.StartTrade.After is True in Buy")
		return
	}

	if !ai.SignalEvents.CanBuy(candle.Time) {
		// log.Println("ai.SignalEvents.CanBuy is False")
		return
	}

	log.Println("CanBuy is True.")
	availableCurrency, _ := ai.GetAvailableBalance()
	useCurrency := availableCurrency * ai.UsePercent
	ticker, err := ai.API.GetTicker(ai.ProductCode)
	if err != nil {
		return
	}
	size := 1 / (ticker.BestAsk / useCurrency)
	size = ai.AdjustSize(size)

	order := &models.Order{
		ProductCode:     ai.ProductCode,
		ChildOrderType:  "MARKET",
		Side:            "BUY",
		Size:            size,
		MinuteToExpires: ai.MinuteToExpires,
		TimeInForce:     "GTC",
	}
	log.Printf("status=buy candle=%+v order=%+v", candle, order)
	childOrderAcceptanceID, err = ai.API.SendOrder(order)
	if err != nil {
		slack.Notice("notification", "Send order failed: " + err.Error())
		log.Println("Send order failed: ", err)
		return
	}
	if childOrderAcceptanceID == "" {
		slack.Notice("notification", "Insufficient fund")
		log.Printf("order=%+v status=no_id", order)
		return
	}
	if config.Config.Exchange == "bitflyer" {
		isOrderCompleted = ai.WaitUntilOrderCompleteBitflyer(childOrderAcceptanceID, candle.Time)
	} else {
		isOrderCompleted = ai.WaitUntilOrderCompleteQuoine(childOrderAcceptanceID, candle.Time)		
	}
	return childOrderAcceptanceID, isOrderCompleted
}

func (ai *AI) Sell(candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool) {
	if ai.BackTest {
		couldSell := ai.SignalEvents.Sell(ai.ProductCode, candle.Time, candle.Close, 1.0, false)
		return "", couldSell
	}

	if ai.StartTrade.After(candle.Time) {
		// if SellToSecureProfit {
		// 	slack.Notice("notification", "ai.StartTrade.After is True in Sell")
		// 	log.Println("ai.StartTrade.After is True in Sell")
		// }
		// log.Println("ai.StartTrade.After is True in Sell")
		return
	}

	if !ai.SignalEvents.CanSell(candle.Time) {
		// if SellToSecureProfit {
		// 	slack.Notice("notification", "ai.SignalEvents.CanSell is False")
		// 	log.Println("ai.SignalEvents.CanSell is False")
		// }
		// log.Println("ai.SignalEvents.CanSell is False")
		return
	}
	// if !SellToSecureProfit {
	// 	log.Println("SellToSecureProfit is true")
	// }

	log.Println("CanSell is True. ")
	_, availableCoin := ai.GetAvailableBalance()
	size := ai.AdjustSize(availableCoin)

	order := &models.Order{
		ProductCode:     ai.ProductCode,
		ChildOrderType:  "MARKET",
		Side:            "SELL",
		Size:            size,
		MinuteToExpires: ai.MinuteToExpires,
		TimeInForce:     "GTC",
	}
	log.Printf("status=sell candle=%+v order=%+v", candle, order)
	childOrderAcceptanceID, err := ai.API.SendOrder(order)
	if err != nil {
		slack.Notice("notification", "Send order failed: " + err.Error())
		log.Println("Send order failed: ", err)
		return
	}
	if childOrderAcceptanceID == "" {
		// Insufficient fund
		slack.Notice("notification", "Insufficient fund")
		log.Printf("order=%+v status=no_id", order)
		return
	}
	if config.Config.Exchange == "bitflyer" {
		isOrderCompleted = ai.WaitUntilOrderCompleteBitflyer(childOrderAcceptanceID, candle.Time)
	} else {
		isOrderCompleted = ai.WaitUntilOrderCompleteQuoine(childOrderAcceptanceID, candle.Time)		
	}
	return childOrderAcceptanceID, isOrderCompleted
}

// 新規に Candle 情報が作成され、なおかつ設定したトレード期間に一致した場合に、
// インディケータのパラメータの最適化と売買判断を実行する。
// streaming.go によって呼び出される
func (ai *AI) Trade(bought_in_same_candle, sold_in_same_candle, is_holding, sellToSecureProfit bool) (bool, bool, bool) {
	isAcquire := ai.TradeSemaphore.TryAcquire(1)
	if !isAcquire {
		slack.Notice("notification", "Could not get trade lock")
		log.Println("Could not get trade lock")
		return bought_in_same_candle, sold_in_same_candle, is_holding
	}
	defer ai.TradeSemaphore.Release(1)
	params := ai.OptimizedTradeParams
	if params == nil {
		log.Println("OptimizedTradeParams is nil.")
		return bought_in_same_candle, sold_in_same_candle, is_holding
	}
	df, _ := models.GetAllCandle(ai.ProductCode, ai.Duration, 100)
	// df, _ := models.GetAllCandle(ai.ProductCode, ai.Duration, ai.PastPeriod)
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
			if emaValues1[i-1] < emaValues2[i-1] && emaValues1[i] > emaValues2[i] {
				buyPoint++
			} else if emaValues1[i] > emaValues2[i] && sold_in_same_candle {
				buyPoint++
			}

			if emaValues1[i-1] > emaValues2[i-1] && emaValues1[i] < emaValues2[i] {
				sellPoint++
			} else if emaValues1[i] < emaValues2[i] && bought_in_same_candle {
				sellPoint++
			}
		}

		if params.BbEnable && params.BbN <= i {
			if bbDown[i-1] > df.Candles[i-1].Close && bbDown[i] < df.Candles[i].Close {
				buyPoint++
			} else if bbDown[i] < df.Candles[i].Close && sold_in_same_candle {
				buyPoint++
			}

			if bbUp[i-1] < df.Candles[i-1].Close && bbUp[i] > df.Candles[i].Close {
				sellPoint++
			} else if bbUp[i] > df.Candles[i].Close && bought_in_same_candle {
				sellPoint++
			}
		}

		if params.MacdEnable {
			if outMACD[i] < 0 && outMACDSignal[i] < 0 && outMACD[i-1] < outMACDSignal[i-1] && outMACD[i] > outMACDSignal[i] {
				buyPoint++
			}

			if outMACD[i] > 0 && outMACDSignal[i] > 0 && outMACD[i-1] > outMACDSignal[i-1] && outMACD[i] < outMACDSignal[i] {
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
			if rsiValues[i-1] < params.RsiBuyThread && rsiValues[i] > params.RsiBuyThread {
				buyPoint++
			} else if rsiValues[i] > params.RsiBuyThread && sold_in_same_candle {
				buyPoint++
			}

			if rsiValues[i-1] > params.RsiSellThread && rsiValues[i] < params.RsiSellThread {
				sellPoint++
			} else if rsiValues[i] < params.RsiSellThread && bought_in_same_candle {
				sellPoint++
			}
		}

		if buyPoint > 0 {
			_, isOrderCompleted := ai.Buy(df.Candles[i])
			if !isOrderCompleted {
				continue
			}
			ai.StopLimit = df.Candles[i].Close * ai.StopLimitPercent
			bought_in_same_candle = true
			sold_in_same_candle = false
			is_holding = true
			BreakEvenPrice = df.Candles[i].Close * config.Config.BreakEvenPercent
			BreakEvenFlagPrice = df.Candles[i].Close * config.Config.BreakEvenFlagPercent
			StreamSellInterval = config.Config.SellInterval
			log.Println("#### StreamSellInterval is ", StreamSellInterval)
			log.Println("#### df.Candles[i].Close is ", df.Candles[i].Close)
			log.Println("#### config.Config.BreakEvenPercent is ", config.Config.BreakEvenPercent)
			log.Println("#### config.Config.BreakEvenFlagPercent is ", config.Config.BreakEvenFlagPercent)
			log.Println("#### BreakEvenPrice is ", BreakEvenPrice)
			log.Println("#### BreakEvenFlagPrice is ", BreakEvenFlagPrice)
			slack.Notice("trade", "BuyPrice : "+ strconv.FormatFloat(df.Candles[i].Close, 'f', 4, 64))			
			slack.Notice("trade", "BreakEvenPrice : "+ strconv.FormatFloat(BreakEvenPrice, 'f', 4, 64) + ", BreakEvenFlagPrice :" + strconv.FormatFloat(BreakEvenFlagPrice, 'f', 4, 64))
			// ProfitConfirmationFlag = false
			// SellToSecureProfit = false
			return bought_in_same_candle, sold_in_same_candle, is_holding
		}

		if sellPoint > 0 || ai.StopLimit > df.Candles[i].Close || sellToSecureProfit {
			_, isOrderCompleted := ai.Sell(df.Candles[i])
			if !isOrderCompleted {
				continue
				}
			
			if sellToSecureProfit{
				log.Println("SellToSecureProfit is excecuted !!!")
				slack.Notice("trade", "SellToSecureProfit is excecuted")
			}

			if ai.StopLimit > df.Candles[i].Close {
				log.Println("### Stop Limit !!!")
				slack.Notice("trade", "Stop Limit !!!")
			}
			ai.StopLimit = 0.0
			ai.UpdateOptimizeParams(true)
			bought_in_same_candle = false
			sold_in_same_candle = true
			is_holding = false
			BreakEvenPrice = 0.0
			BreakEvenFlagPrice = 0.0
			ProfitConfirmationFlag = false
			// SellToSecureProfit = false
			return bought_in_same_candle, sold_in_same_candle, is_holding
		}

	}
	return bought_in_same_candle, sold_in_same_candle, is_holding
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
	// fee := size * ApiFeePercent
	// size = size - fee
	return math.Floor(size*10000) / 10000
}

// For BITFLYER
func (ai *AI) WaitUntilOrderCompleteBitflyer(childOrderAcceptanceID string, executeTime time.Time) bool {
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

// For QUOINE
func (ai *AI) WaitUntilOrderCompleteQuoine(orderID string, executeTime time.Time) bool {
	params := map[string]string{
		// "product_code":              ai.ProductCode,
		// "child_order_acceptance_id": childOrderAcceptanceID,
		"orderID": orderID,
	}
	expire := time.After(time.Minute + (20 * time.Second))
	interval := time.Tick(5 * time.Second)
	return func() bool {
		for {
			select {
			case <-expire:
				slack.Notice("trade", "Expire.")
				return false
			case <-interval:
				listOrders, err := ai.API.ListOrder(params)
				if err != nil {
					log.Println(err)
					slack.Notice("trade", "Get List Order failed.")
					return false
				}
				if len(listOrders) == 0 {
					slack.Notice("trade", "len(listOrders) == 0 ")
					return false
				}
				order := listOrders[0]
				if order.Status == "filled" {
					if order.Side == "buy" {
						tradedPrice := order.Price * order.FilledQuantity
						couldBuy := ai.SignalEvents.Buy(ai.ProductCode, executeTime, tradedPrice, order.FilledQuantity, true)
						// if !couldBuy {
						if couldBuy {
							strTradePrice := strconv.FormatFloat(tradedPrice, 'f', 4, 64)
							slack.Notice("trade", "BUY process completed ! ==> " + strTradePrice)
							log.Printf("status=buy orderID=%s order=%+v", orderID, order)
							var indicator string
							var param1, param2, param3 float64
							indicator, param1, param2, param3 = ai.GetIndicatorInfo()
							models.InsertBuyResult(executeTime, tradedPrice, order.Price, config.Config.StopLimitPercent,
								param1, param2, param3, indicator, config.Config.Exchange, config.Config.ProductCode,
								os.Getenv("TRADE_DURATION"), os.Getenv("REFERENCE_DURATION1"), os.Getenv("REFERENCE_DURATION2"),
								config.Config.DataLimit, config.Config.NumRanking)
						}
						return couldBuy
					}
					if order.Side == "sell" {
						tradedPrice := order.Price * order.FilledQuantity
						couldSell := ai.SignalEvents.Sell(ai.ProductCode, executeTime, tradedPrice, order.FilledQuantity, true)
						// if !couldSell {
						if couldSell {
							strTradePrice := strconv.FormatFloat(tradedPrice, 'f', 4, 64)
							slack.Notice("trade", "SELL process completed ! ==> " + strTradePrice)
							log.Printf("status=sell orderID=%s order=%+v", orderID, order)
							var balance float64
							balance, _ = ai.GetAvailableBalance()
							models.UpdateSellResult(executeTime, tradedPrice, balance, order.Price)
						}
						return couldSell
					}
					return false
				}
				// ChildOrderState == "COMPLETED" {
				// 	if order.Side == "BUY" {
				// 		couldBuy := ai.SignalEvents.Buy(ai.ProductCode, executeTime, order.AveragePrice, order.Size, true)
				// 		if !couldBuy {
				// 			slack.Notice("trade", "BUY process completed !")
				// 			log.Printf("status=buy childOrderAcceptanceID=%s order=%+v", childOrderAcceptanceID, order)
				// 		}
				// 		return couldBuy
				// 	}
				// 	if order.Side == "SELL" {
				// 		couldSell := ai.SignalEvents.Sell(ai.ProductCode, executeTime, order.AveragePrice, order.Size, true)
				// 		if !couldSell {
				// 			slack.Notice("trade", "SELL process completed !")
				// 			log.Printf("status=sell childOrderAcceptanceID=%s order=%+v", childOrderAcceptanceID, order)
				// 		}
				// 		return couldSell
				// 	}
				// 	return false
				// }
			}
		}
	}()
}