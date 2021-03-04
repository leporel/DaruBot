package candles

import (
	"github.com/markcheno/go-quote"
	"time"
)

var (
	//testPair = "BTC-USD" // coinbase
	testPair = "BTCUSDT" // binance-usdt

	//market = "coinbase"
	//quoteFrom = quote.NewQuoteFromCoinbase
	market    = "binance-usdt"
	quoteFrom = quote.NewQuoteFromBinance

	timeFormat = "2006-01-02 15:04"
)

func quoteFormat(t time.Time) string {
	return t.Format(timeFormat)
}

func downloadQuote(from, to time.Time, symbol string, period quote.Period) (*quote.Quote, error) {
	start := from.UTC()
	end := to.UTC()
	q, err := quoteFrom(symbol, quoteFormat(start), quoteFormat(end), period)
	if err != nil {
		return nil, err
	}
	return &q, nil
}
