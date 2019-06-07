package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type ConfigList struct {
	ApiKey      string
	ApiSecret   string
	LogFile     string
	ProductCode string

	TradeDuration time.Duration
	Durations     map[string]time.Duration
	DbName        string
	SQLDriver     string
	Port          int

	BackTest         bool
	UsePercent       float64
	DataLimit        int
	StopLimitPercent float64
	NumRanking       int
	Exchange         string
}

var Config ConfigList

func Env_load() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func init() {
	Env_load()

	durations := map[string]time.Duration{
		"1s": time.Second,
		"1m": time.Minute,
		"1h": time.Hour,
		"1d": time.Hour * 24,
	}

	var apiKey, apiSecret string

	if os.Getenv("EXCHANGE") == "bitflyer" {
		apiKey = os.Getenv("BITFLYER_API_KEY")
		apiSecret = os.Getenv("BITFLYER_API_SECRET")
	} else if os.Getenv("EXCHANGE") == "quoine" {
		apiKey = os.Getenv("QUOINE_API_KEY")
		apiSecret = os.Getenv("QUOINE_API_SECERT")
	}

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	backTest, _ := strconv.ParseBool(os.Getenv("BACK_TEST"))
	usePercent, _ := strconv.ParseFloat(os.Getenv("USE_PERCENT"), 64)
	dataLimit, _ := strconv.Atoi(os.Getenv("DATA_LIMIT"))
	stopLimitPercent, _ := strconv.ParseFloat(os.Getenv("STOP_LIMIT_PERCENT"), 64)
	numRanking, _ := strconv.Atoi(os.Getenv("NUM_RANKING"))

	Config = ConfigList{
		ApiKey:           apiKey,
		ApiSecret:        apiSecret,
		LogFile:          os.Getenv("LOG_FILE"),
		ProductCode:      os.Getenv("PRODUCT_CODE"),
		Durations:        durations,
		TradeDuration:    durations[os.Getenv("TRADE_DURATION")],
		DbName:           os.Getenv("DATABASE_URL"),
		SQLDriver:        os.Getenv("DATABASE_DRIVER"),
		Port:             port,
		BackTest:         backTest,
		UsePercent:       usePercent,
		DataLimit:        dataLimit,
		StopLimitPercent: stopLimitPercent,
		NumRanking:       numRanking,
		Exchange:         os.Getenv("EXCHANGE"),
	}

	// Existing Code
	// cfg, err := ini.Load("config.ini")
	// if err != nil {
	// 	log.Printf("Failed to read file: %v", err)
	// 	os.Exit(1)
	// }

	// durations := map[string]time.Duration{
	// 	"1s": time.Second,
	// 	"1m": time.Minute,
	// 	"1h": time.Hour,
	// 	"1d": time.Hour * 24,
	// }

	// var apiKey, apiSecret string

	// if cfg.Section("gotrading").Key("exchange").String() == "bitflyer" {
	// 	apiKey = cfg.Section("bitflyer").Key("api_key").String()
	// 	apiSecret = cfg.Section("bitflyer").Key("api_secret").String()
	// } else if cfg.Section("gotrading").Key("exchange").String() == "quoine" {
	// 	apiKey = cfg.Section("quoine").Key("api_key").String()
	// 	apiSecret = cfg.Section("quoine").Key("api_secret").String()
	// }

	// port_str := os.Getenv("PORT")
	// var port int
	// if port_str == "" {
	// 	port = cfg.Section("web").Key("port").MustInt()
	// } else {
	// 	port, _ = strconv.Atoi(port_str)
	// }

	// var dbName string
	// if os.Getenv("DATABASE_URL") == "" {
	// 	dbName = cfg.Section("db").Key("name").String()
	// } else {
	// 	// TODO
	// 	// Heroku で作成される MySQL のエンドポイントを指定する。
	// 	dbName = ""
	// }

	// Config = ConfigList{
	// 	ApiKey:           apiKey,
	// 	ApiSecret:        apiSecret,
	// 	LogFile:          cfg.Section("gotrading").Key("log_file").String(),
	// 	ProductCode:      cfg.Section("gotrading").Key("product_code").String(),
	// 	Durations:        durations,
	// 	TradeDuration:    durations[cfg.Section("gotrading").Key("trade_duration").String()],
	// 	DbName:           dbName,
	// 	SQLDriver:        cfg.Section("db").Key("driver").String(),
	// 	Port:             port,
	// 	BackTest:         cfg.Section("gotrading").Key("back_test").MustBool(),
	// 	UsePercent:       cfg.Section("gotrading").Key("use_percent").MustFloat64(),
	// 	DataLimit:        cfg.Section("gotrading").Key("data_limit").MustInt(),
	// 	StopLimitPercent: cfg.Section("gotrading").Key("stop_limit_percent").MustFloat64(),
	// 	NumRanking:       cfg.Section("gotrading").Key("num_ranking").MustInt(),
	// 	Exchange:         cfg.Section("gotrading").Key("exchange").String(),
	// }
}
