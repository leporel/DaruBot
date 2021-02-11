package models

import (
	"strconv"
	"time"
)

type OrderType string

const (
	OrderTypeLimit     OrderType = "LIMIT"
	OrderTypeMarket    OrderType = "MARKET"
	OrderTypeStop      OrderType = "STOP"
	OrderTypeStopLimit OrderType = "STOP LIMIT"
	OrderTypeUnknown   OrderType = "UNKNOWN"
)

type PutOrder struct {
	InternalID string
	Pair       string
	Type       OrderType
	Amount     float64
	// Positive for buy, Negative for sell, ignoring if OrderTypeMarket
	Price     float64
	StopPrice float64

	Margin bool
}

type Order struct {
	ID             string
	InternalID     string
	Type           OrderType
	Price          float64
	PriceAvg       float64
	AmountCurrent  float64
	AmountOriginal float64
	Date           time.Time
	Updated        time.Time
	Meta           map[string]interface{}
}

func (o *Order) GetID() string {
	return o.ID
}

func (o *Order) GetIDAsInt() int64 {
	if o.ID == "" {
		return 0
	}

	id, err := strconv.ParseInt(o.ID, 10, 64)
	if err != nil {
		panic(err)
	}
	return id
}

func (o *Order) GetInternalID() string {
	return o.InternalID
}

func (o *Order) GetInternalIDAsInt() int64 {
	if o.InternalID == "" {
		return 0
	}

	id, err := strconv.ParseInt(o.InternalID, 10, 64)
	if err != nil {
		panic(err)
	}
	return id
}

func (o *Order) GetPrice() float64 {
	return o.Price
}

func (o *Order) GetAmount() float64 {
	return o.AmountCurrent
}

func (o *Order) GetOriginalAmount() float64 {
	return o.AmountOriginal
}

func (o *Order) GetType() OrderType {
	return o.Type
}

func (o *Order) IsFilled() bool {
	return o.AmountCurrent == 0
}
