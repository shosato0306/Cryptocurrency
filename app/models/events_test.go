package models

import (
	"testing"
	"log"
	"time"
)

func TestCanBuy(t *testing.T) {
	s := GetSignalEventsByCount(1)
	// signalEvent := &Signa
	log.Printf("%+v", s)
	requestTime, err := time.Parse("2006-01-02 15:04:05", "2019-06-10 22:15:11")
	if err != nil {
		t.Fatal(err)
	}
	canBuyResult := s.CanBuy(requestTime)
	log.Println("### Can BUY")
	log.Println(canBuyResult)
	canSellResult := s.CanSell(requestTime)
	log.Println("### Can SELL")
	log.Println(canSellResult)
}
