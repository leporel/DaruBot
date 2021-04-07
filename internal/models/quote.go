package models

import (
	"fmt"
	"github.com/markcheno/go-quote"
)

func (r *CandleResolution) ToQuoteModel() (quote.Period, error) {
	switch *r {
	case OneMinute:
		return quote.Min1, nil
	case FiveMinutes:
		return quote.Min5, nil
	case FifteenMinutes:
		return quote.Min15, nil
	case ThirtyMinutes:
		return quote.Min30, nil
	case OneHour:
		return quote.Min60, nil
	case SixHours:
		return quote.Hour6, nil
	case TwelveHours:
		return quote.Hour12, nil
	case OneDay:
		return quote.Daily, nil
	case OneWeek:
		return quote.Weekly, nil
	case OneMonth:
		return quote.Monthly, nil
	default:
		return "", fmt.Errorf("period (⊙_⊙)？")
	}
}

func QuoteToModels(q *quote.Quote, symbol string) *Candles {
	var r CandleResolution
	if len(q.Date) > 1 {
		r, _ = candleResolutionFromDuration(q.Date[1].Sub(q.Date[0]))
	}

	rs := &Candles{
		Symbol:     symbol,
		Resolution: r,
		Candles:    make([]*Candle, 0, len(q.Date)),
	}

	for i, _ := range q.Date {
		rs.Candles = append(rs.Candles, QuoteToModel(q, symbol, i, r))
	}

	return rs
}

func QuoteToModel(q *quote.Quote, symbol string, index int, r CandleResolution) *Candle {
	if r == "" && len(q.Date) > 1 {
		r, _ = candleResolutionFromDuration(q.Date[1].Sub(q.Date[0]))
	}

	return &Candle{
		Symbol:     symbol,
		Resolution: r,
		Date:       q.Date[index],
		Open:       q.Open[index],
		Close:      q.Close[index],
		High:       q.High[index],
		Low:        q.Low[index],
		Volume:     q.Volume[index],
	}
}