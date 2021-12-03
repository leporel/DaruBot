package models

import "strconv"

type Position struct {
	ID                   string
	Symbol               string
	Price                float64
	Amount               float64
	LiqPrice             float64
	MarginLevel          float64
	ProfitLoss           float64
	ProfitLossPercentage float64
	Meta                 map[string]interface{}
}

func (p *Position) GetID() string {
	return p.ID
}

func (p *Position) GetIDAsInt() int64 {
	if p.ID == "" {
		return 0
	}
	id, err := strconv.ParseInt(p.ID, 10, 64)
	if err != nil {
		panic(err)
	}
	return id
}

func (p *Position) GetPrice() float64 {
	return p.Price
}

func (p *Position) GetAmount() float64 {
	return p.Amount
}

func (p *Position) GetLiquidationPrice() float64 {
	return p.LiqPrice
}

func (p *Position) GetMarginLevel() float64 {
	return p.MarginLevel
}

func (p *Position) GetProfit() float64 {
	return p.ProfitLoss
}

func (p *Position) GetProfitPercentage() float64 {
	return p.ProfitLossPercentage
}
