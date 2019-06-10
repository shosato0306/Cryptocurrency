package quoine

import (
	"bytes"
	"cryptocurrency/slack"
	"cryptocurrency/bitflyer"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"fmt"

	"cryptocurrency/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/toorop/go-pusher"
)

const baseURL = "https://api.liquid.com/"

type APIClient struct {
	key        string
	secret     string
	httpClient *http.Client
}

func New(key, secret string) *APIClient {
	apiClient := &APIClient{key, secret, &http.Client{}}
	return apiClient
}

// // header リクエストに追加するヘッダー情報
func (api APIClient) header(method, endpoint string, body []byte) map[string]string {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims = jwt.MapClaims{
		"path":     endpoint,
		"nonce":    strconv.FormatInt(time.Now().Unix(), 10),
		"token_id": config.Config.ApiKey,
	}
	signature, _ := token.SignedString([]byte(config.Config.ApiSecret))
	return map[string]string{
		"X-Quoine-Auth":        signature,
		"X-Quoine-API-Version": "2",
		"Content-Type":         "application/json",
	}
}

type Product struct {
	ID                  json.Number     `json:"id"`
	ProductType         string  `json:"product_type"`
	Code                string  `json:"code"`
	Name                string  `json:"name"`
	MarketAsk           float64 `json:"market_ask,string"`
	MarketBid           float64 `json:"market_bid,string"`
	Indicator           int     `json:"indicator"`
	Currency            string  `json:"currency"`
	CurrencyPairCode    string  `json:"currency_pair_code"`
	Symbol              string  `json:"symbol"`
	BtcMinimumWithdraw  float64 `json:"btc_minimum_withdraw,string,omitempty"`
	FiatMinimumWithdraw float64 `json:"fiat_minimum_withdraw,string,omitempty"`
	PusherChannel       string  `json:"pusher_channel"`
	LowMarketBid        float64 `json:"low_market_bid,string"`
	HighMarketAsk       float64 `json:"high_market_ask,string"`
	Volume24H           float64 `json:"volume_24h,string"`
	LastPrice24H        float64 `json:"last_price_24h,string"`
	LastTradedPrice     float64 `json:"last_traded_price,string"`
	LastTradedQuantity  float64 `json:"last_traded_quantity,string"`
	QuotedCurrency      string  `json:"quoted_currency"`
	BaseCurrency        string  `json:"base_currency"`
	Disabled            bool    `json:"disabled"`
	MarginEnabled       bool    `json:"margin_enabled"`
	CfdEnabled          bool    `json:"cfd_enabled"`
	LastEventTimestamp  string  `json:"last_event_timestamp"`
}

const APP_KEY = "2ff981bb060680b5ce97"

func (api *APIClient) GetRealTimeProduct(symbol string, ch chan<- *bitflyer.Ticker) {
INIT:
	log.Println("init...")

	pusherClient, err := pusher.NewClient(APP_KEY)
	if err != nil {
		slack.Notice("notification", "Create NewClient failed: " + err.Error())
		log.Println(err)
		log.Println("wait...")
		time.Sleep(time.Second * 5)
		goto INIT
	}

	// Subscribe
	err = pusherClient.Subscribe("product_cash_btcjpy_5")
	if err != nil {
		slack.Notice("notification", "Subscribe failed: " + err.Error())
		log.Println("Subscription error : ", err)
	}

	updatedChannel, err := pusherClient.Bind("updated")
	if err != nil {
		slack.Notice("notification", "Bind failed: " + err.Error())
		log.Println("Bind error: ", err)
	}
	log.Println("Binded to 'update' event")

	errChannel, err := pusherClient.Bind(pusher.ErrEvent)
	if err != nil {
		slack.Notice("notification", "Bind failed: " + err.Error())
		log.Println("Bind error: ", err)
	}
	log.Println("Binded to 'ErrEvent' event")

	log.Println("init done")

	slack.Notice("notification", "Initialization of connection to quine is complete")

	for {
		select {
		case dataEvt := <-updatedChannel:
			bytes := []byte(dataEvt.Data)
			var product Product
			err = json.Unmarshal(bytes, &product)
			if err != nil {
				slack.Notice("notification", "Unmarshal failed: " + err.Error())
				log.Fatal(err)
			}

			productCode := product.BaseCurrency + "_" + product.Currency

			eventTimestamp := strings.Split(product.LastEventTimestamp, ".")[0]
			intEventTimestamp, _ := strconv.ParseInt(eventTimestamp, 10, 64)
			strEventTimestamp := time.Unix(intEventTimestamp, 0).UTC().Format(time.RFC3339)
			Ticker := bitflyer.NewTicker(productCode, strEventTimestamp, product.MarketBid, product.MarketAsk, product.Volume24H)

			ch <- Ticker

		case errEvt := <-errChannel:
			log.Println("ErrEvent: " + errEvt.Data)
			pusherClient.Close()
			time.Sleep(time.Second)
			goto INIT
		}
	}
}

func (api *APIClient) doRequest(method, urlPath string, query map[string]string, data []byte) (body []byte, err error) {
	baseURL, err := url.Parse(baseURL)
	if err != nil {
		return
	}
	apiURL, err := url.Parse(urlPath)
	if err != nil {
		return
	}
	endpoint := baseURL.ResolveReference(apiURL).String()
	log.Printf("action=doRequest endpoint=%s", endpoint)
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	for key, value := range api.header(method, req.URL.RequestURI(), data) {
		req.Header.Add(key, value)
	}
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type Balances []struct {
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
}

func (api *APIClient) GetBalances() (Balances, error){
	url := "/accounts/balance"
	var balances Balances
	resp, err := api.doRequest("GET", url, nil, nil)
	if err != nil {
		log.Printf("action=GETBalances err=%s", err.Error())
		return balances, err
	}
	err = json.Unmarshal(resp, &balances)
	if err != nil {
		log.Printf("action=GETBalances(unmarshal) err=%s", err.Error())
		return balances, err
	}
	return balances, err
}

func (api *APIClient) GetProduct() (Product, error) {
	url := "/products/5"
	var product Product
	resp, err := api.doRequest("GET", url, nil, nil)
	if err != nil {
		log.Printf("action=GetProduct err=%s", err.Error())
		return product, err
	}
	err = json.Unmarshal(resp, &product)
	if err != nil {
		log.Printf("action=GetProduct(unmarshal) err=%s", err.Error())
		return product, err
	}
	return product, err
}

func (api *APIClient) GetExecutions() {
	url := "/executions/me?product_id=5"
	// var product Product
	resp, err := api.doRequest("GET", url, nil, nil)
	log.Println(string(resp))
	if err != nil {
		log.Printf("action=GetExecutions err=%s", err.Error())
		// return product
	}
	// err = json.Unmarshal(resp, &product)
	// if err != nil {
	// 	log.Printf("action=GetOrders(unmarshal) err=%s", err.Error())
		// return product
	// }
	// return product
}

type Order struct {
	OrderDetail  OrderDetail `json:"order"`
}

type OrderDetail struct {
	OrderType string `json:"order_type"`
	ProductID int    `json:"product_id"`
	Side      string `json:"side"`
	Quantity  string `json:"quantity"`
}

type ResponseSendChildOrder struct {
    ID                   int         `json:"id"`
    OrderType            string      `json:"order_type"`
    Quantity             float64     `json:"quantity,string"`
    DiscQuantity         string      `json:"disc_quantity"`
    IcebergTotalQuantity string      `json:"iceberg_total_quantity"`
    Side                 string      `json:"side"`
    FilledQuantity       float64     `json:"filled_quantity,string"`
    Price                float64     `json:"price"`
    CreatedAt            int         `json:"created_at"`
    UpdatedAt            int         `json:"updated_at"`
    Status               string      `json:"status"`
    LeverageLevel        int         `json:"leverage_level"`
    SourceExchange       string      `json:"source_exchange"`
    ProductID            int         `json:"product_id"`
    ProductCode          string      `json:"product_code"`
    FundingCurrency      string      `json:"funding_currency"`
    CryptoAccountID      interface{} `json:"crypto_account_id"`
    CurrencyPairCode     string      `json:"currency_pair_code"`
    AveragePrice         float64     `json:"average_price"`
    Target               string      `json:"target"`
    OrderFee             float64     `json:"order_fee"`
    SourceAction         string      `json:"source_action"`
    UnwoundTradeID       interface{} `json:"unwound_trade_id"`
    TradeID              interface{} `json:"trade_id"`
}


func (api *APIClient) SendOrder(side, quantity string) (*ResponseSendChildOrder, error) {
	var order Order
	order = Order{
		OrderDetail: OrderDetail {
			OrderType: "market",
			ProductID: 5, 
			Side: side, 
			Quantity: quantity, 
		},
	}

	data, _ := json.Marshal(order)
	fmt.Println(string(data))
    url := "/orders/"
	resp, err := api.doRequest("POST", url, map[string]string{}, data)
    if err != nil {
        log.Printf("Order Request fail, err=%s", err.Error())
        return nil, err
    }
    var response ResponseSendChildOrder
    err = json.Unmarshal(resp, &response)
    if err != nil {
        log.Printf("Order Request Unmarshal fail, err=%s", err.Error())
        return nil, err
    }
    return &response, nil
}

// // GetOrder 注文IDの情報を取得する
// func (api *APIClient) GetOrder(orderID int) (*Order, error) {
//     var getOrder *Order
//     // spath := fmt.Sprintf("/orders/%d", orderID)
//     spath := fmt.Sprintf("/orders?with_details=1")
// 	resp, err := api.doRequest("GET", spath, nil, nil)
// 	log.Println(string(resp))	
//     if err != nil {
//         log.Printf("Get Order Request Error, err = %s", err.Error())
//         return nil, err
//     }

//     err = json.Unmarshal(resp, &getOrder)
//     if err != nil {
//         log.Printf("Get Order Request Unmarshal Error, err = %s", err.Error())
//         return nil, err
//     }
//     return getOrder, nil
// }


// 売値と買値の中間の値を取得
func (p *Product) GetMidPrice() float64 {
	return (p.MarketBid + p.MarketAsk) / 2
}

func (p *Product) DateTime() time.Time {
	// LastEventTimestamp が適当な値かは確認が必要
	dateTime, err := time.Parse(time.RFC3339, p.LastEventTimestamp)
	if err != nil {
		log.Printf("action=DateTime, err=%s", err.Error())
	}
	return dateTime
}

func (p *Product) TruncateDateTime(duration time.Duration) time.Time {
	return p.DateTime().Truncate(duration)
}
