package models

import "DaruBot/internal/models/exchanges"

type Ticker struct {
	Symbol   string
	Price    float64
	Exchange exchanges.ExchangeType
	State    TickerState
}

// Ticker https://docs.bitfinex.com/reference#rest-public-tickers
type TickerState struct {
	High    float64
	Low     float64
	Volume  float64
	BID     float64
	BIDSize float64
	ASK     float64
	ASKSize float64
}
