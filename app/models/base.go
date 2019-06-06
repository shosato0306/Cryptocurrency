package models

import (
	"cryptocurrency/config"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/mattn/go-sqlite3"
)

const (
	tableNameSignalEvents = "signal_events"
)

var DbConnection *sql.DB
var GormDbConnection *gorm.DB
var DB *sql.DB

func GetCandleTableName(productCode string, duration time.Duration) string {
	return fmt.Sprintf("%s_%s", productCode, duration)
}

func init() {
	var err error
	DbConnection, err = sql.Open(config.Config.SQLDriver, config.Config.DbName)
	if err != nil {
		log.Fatalln(err)
	}
	cmd := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            time DATETIME PRIMARY KEY NOT NULL,
            product_code STRING,
            side STRING,
            price FLOAT,
            size FLOAT)`, tableNameSignalEvents)
	DbConnection.Exec(cmd)

	for _, duration := range config.Config.Durations {
		tableName := GetCandleTableName(config.Config.ProductCode, duration)
		c := fmt.Sprintf(`
            CREATE TABLE IF NOT EXISTS %s (
            time DATETIME PRIMARY KEY NOT NULL,
            open FLOAT,
            close FLOAT,
            high FLOAT,
            low open FLOAT,
			volume FLOAT)`, tableName)
		DbConnection.Exec(c)
	}

	GormDbConnection, err = gorm.Open("mysql", "root:@/cryptocurrency?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("DB connection error: ", err)
	}

	for _, duration := range config.Config.Durations {
		tableName := GetCandleTableName(config.Config.ProductCode, duration)
		CreatedTableName = tableName

		GormDbConnection.AutoMigrate(&Candle{})
	}

	GormDbConnection.AutoMigrate(&SignalEvent{})

	DB, err = sql.Open("mysql", "root:@/cryptocurrency?charset=utf8&parseTime=True&loc=Local")
}

var CreatedTableName string

func (c *Candle) TableName() string {
	return CreatedTableName
}
