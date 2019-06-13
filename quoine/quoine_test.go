package quoine

import (
	"time"
	"testing"
	"log"
	"cryptocurrency/config"
	"cryptocurrency/app/models"
)

func TestGetBalance(t *testing.T) {
	t.Skip("... skip TestGetBalances")
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	balances, err := apiClient.GetBalance()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("balances => %+v", balances)
}

func TestGetTicker(t *testing.T) {
	t.Skip("... skip TestGetProduct")
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	ticker, err := apiClient.GetTicker("Dummy string")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("ticker => %+v", ticker)
}

// func TestGetExecutions(t *testing.T) {
// 	t.Skip("... skip TestGetExecutions")
// 	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
// 	apiClient.GetExecutions()
// }

func TestListOrder(t *testing.T) {
	t.Skip("... skip TestGetOrders")
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	params := map[string]string{"orderID":"1137604402"}
	order, err := apiClient.ListOrder(params)
	if err != nil {
		t.Error()
	}
	log.Printf("%+v", order)
	log.Println(order[0].Status)
	log.Println(order[0].Side)
	log.Println(order[0].FilledQuantity)
}

func TestSendOrder(t *testing.T) {
	t.Skip("... skip TestSendOrder")
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	order := &models.Order{
		ProductCode:     "BTC_JPY",
		ChildOrderType:  "MARKET",
		Side:            "SELL", // 小文字でなければならない！！
		Size:            0.001,
		MinuteToExpires: 1,
		TimeInForce:     "GTC",
	}
	_, err := apiClient.SendOrder(order)
	if err != nil {
		t.Fatal(err)
	}
	// if response == "0" {
	// 	slack.Notice("notification",  "Insufficient fund")
	// } 
	// else {
	// 	slack.Notice("trade", "Trade completed!! Side: " +  response.Side + ", Price: " + strconv.FormatFloat(response.Price * response.FilledQuantity, 'f', 0, 64))
	// }
}

func TestGetRealTimeProduct(t *testing.T) {
	// t.Skip("... skip TestGetRealTimeProduct")
	var tickerChannel = make(chan *models.Ticker)
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	go apiClient.GetRealTimeProduct(config.Config.ProductCode, tickerChannel)
	go func() {
		for ticker := range tickerChannel {
			// for _, duration := range config.Config.Durations {
				// isCreated := models.CreateCandleWithDuration(*ticker, ticker.ProductCode, duration)
				// _ = models.CreateCandleWithDuration(*ticker, ticker.ProductCode, duration)
				// if isCreated == true && duration == config.Config.TradeDuration {
				// if duration == config.Config.TradeDuration {
					// log.Println("### Trade() is called")
					// is_during_buy := false
					// is_ordered = ai.Trade()
					// ai.Trade()
					// if call_count >= 5 && is_during_buy != false {
					// 	ai.UpdateOptimizeParams(true)
					// }
				// }		
			// }
		log.Printf("ticker => %+v", ticker)
		}
	}()
	time.Sleep(20 * time.Second)
}