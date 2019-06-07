package models

import (
	"cryptocurrency/config"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	// _ "github.com/mattn/go-sqlite3"
)

const (
	tableNameSignalEvents = "signal_events"
)

// var GormDbConnection *gorm.DB
var DB *sql.DB

func GetCandleTableName(productCode string, duration time.Duration) string {
	return fmt.Sprintf("%s_%s", productCode, duration)
}

func init() {
	var err error
	DB, err = sql.Open(config.Config.SQLDriver, config.Config.DbName)
	if err != nil {
		log.Fatal("DB connection error: ", err)
	}

	cmd := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            time TIMESTAMP PRIMARY KEY NOT NULL,
            product_code VARCHAR(255),
            side VARCHAR(255),
            price DOUBLE,
            size DOUBLE);`, tableNameSignalEvents)
	DB.Exec(cmd)

	for _, duration := range config.Config.Durations {
		tableName := GetCandleTableName(config.Config.ProductCode, duration)
		c := fmt.Sprintf(`
            CREATE TABLE IF NOT EXISTS %s (
            time TIMESTAMP PRIMARY KEY NOT NULL,
            open DOUBLE,
            close DOUBLE,
            high DOUBLE,
            low DOUBLE,
			volume DOUBLE);`, tableName)
		DB.Exec(c)
	}

	// GormDbConnection, err = gorm.Open(config.Config.SQLDriver, config.Config.DbName)
	// if err != nil {
	// 	log.Fatal("Gorm DB connection error: ", err)
	// }

	// for _, duration := range config.Config.Durations {
	// 	tableName := GetCandleTableName(config.Config.ProductCode, duration)
	// 	CreatedTableName = tableName

	// 	GormDbConnection.AutoMigrate(&Candle{})
	// }

	// GormDbConnection.AutoMigrate(&SignalEvent{})
}

// var CreatedTableName string

// func (c *Candle) TableName() string {
// 	return CreatedTableName
// }
