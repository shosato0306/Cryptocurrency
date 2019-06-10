package quoine

import (
	"testing"
	"log"
	"strconv"
	"cryptocurrency/config"
	"cryptocurrency/slack"
)

func TestGetBalances(t *testing.T) {
	// t.Skip("... skip TestGetBalances")
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	balances, err := apiClient.GetBalances()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("balances => %+v", balances)
}

func TestGetProduct(t *testing.T) {
	// t.Skip("... skip TestGetProduct")
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	product, err := apiClient.GetProduct()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("product => %+v", product)
}

func TestGetExecutions(t *testing.T) {
	// t.Skip("... skip TestGetExecutions")
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	apiClient.GetExecutions()
}

// func TestGetOrder(t *testing.T) {
// 	// t.Skip("... skip TestGetOrders")
// 	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
// 	order, err := apiClient.GetOrder(132349532)
// 	if err != nil {
// 		t.Error()
// 	}
// 	fmt.Printf("%+v", order)
// }

func TestSendOrder(t *testing.T) {
	t.Skip("... skip TestSendOrder")
	apiClient := New(config.Config.ApiKey, config.Config.ApiSecret)
	response, err := apiClient.SendOrder("buy", "0.001")
	if err != nil {
		t.Fatal(err)
	}
	if response.ID == 0 {
		slack.Notice("notification",  "Insufficient fund")
	} else {
		slack.Notice("trade", "Trade completed!! Side: " +  response.Side + ", Price: " + strconv.FormatFloat(response.Price * response.FilledQuantity, 'f', 0, 64))
	}
}