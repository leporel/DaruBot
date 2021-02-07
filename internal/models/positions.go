package models

type Position struct {
	ID                   string
	Pair                 string
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