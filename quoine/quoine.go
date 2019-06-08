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
		"token_id": "APIキーをここに",
	}
	signature, _ := token.SignedString([]byte("APIシークレットをここに"))
	return map[string]string{
		"X-Quoine-Auth":        signature,
		"X-Quoine-API-Version": "2",
		"Content-Type":         "application/json",
	}
}

type Product struct {
	ID                  int     `json:"id"`
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
		slack.Notice("notification", "NewClient failed: " + err.Error())
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

// // TODO
// type Balance struct {
// 	Currency string `json:"currency"`
// 	Balance  string `json:"balance"`
// }

// // GetBalances 現在の総合資産を取得する
// func (api *APIClient) GetBalances() []Balance {
// 	url := "/accounts/balance"
// 	var balances []Balance
// 	resp, err := api.doRequest("GET", url, nil, nil)
// 	if err != nil {
// 		log.Printf("action=GETBalances err=%s", err.Error())
// 		return balances
// 	}
// 	err = json.Unmarshal(resp, &balances)
// 	if err != nil {
// 		log.Printf("action=GETBalances(unmarshal) err=%s", err.Error())
// 		return balances
// 	}
// 	return balances
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
