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

var GormDbConnection *gorm.DB
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
	GormDbConnection, err = gorm.Open(config.Config.SQLDriver, config.Config.DbName)
	if err != nil {
		log.Fatal("Gorm DB connection error: ", err)
	}

	for _, duration := range config.Config.Durations {
		tableName := GetCandleTableName(config.Config.ProductCode, duration)
		CreatedTableName = tableName

		GormDbConnection.AutoMigrate(&Candle{})
	}

	GormDbConnection.AutoMigrate(&SignalEvent{})
}

var CreatedTableName string

func (c *Candle) TableName() string {
	return CreatedTableName
}
