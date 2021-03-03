package models

import (
	"fmt"
	"github.com/markcheno/go-quote"
	"time"
)

type CandleResolution string

const (
	OneMinute      CandleResolution = "1m"
	FiveMinutes    CandleResolution = "5m"
	FifteenMinutes CandleResolution = "15m"
	ThirtyMinutes  CandleResolution = "30m"
	OneHour        CandleResolution = "1h"
	ThreeHours     CandleResolution = "3h"
	SixHours       CandleResolution = "6h"
	TwelveHours    CandleResolution = "12h"
	OneDay         CandleResolution = "D"
	OneWeek        CandleResolution = "W"
	OneMonth       CandleResolution = "M"
)

func CandleResolutionFromString(str string) (CandleResolution, error) {
	switch str {
	case string(OneMinute):
		return OneMinute, nil
	case string(FiveMinutes):
		return FiveMinutes, nil
	case string(FifteenMinutes):
		return FifteenMinutes, nil
	case string(ThirtyMinutes):
		return ThirtyMinutes, nil
	case string(OneHour):
		return OneHour, nil
	case string(ThreeHours):
		return ThreeHours, nil
	case string(SixHours):
		return SixHours, nil
	case string(TwelveHours):
		return TwelveHours, nil
	case string(OneDay):
		return OneDay, nil
	case string(OneWeek):
		return OneWeek, nil
	case string(OneMonth):
		return OneMonth, nil
	}
	return OneMinute, fmt.Errorf("could not convert string to resolution: %s", str)
}

func candleResolutionFromDuration(d time.Duration) (CandleResolution, error) {
	switch d {
	case time.Minute:
		return OneMinute, nil
	case time.Minute * 5:
		return FiveMinutes, nil
	case time.Minute * 15:
		return FifteenMinutes, nil
	case time.Minute * 30:
		return ThirtyMinutes, nil
	case time.Hour:
		return OneHour, nil
	case time.Hour * 3:
		return ThreeHours, nil
	case time.Hour * 6:
		return SixHours, nil
	case time.Hour * 12:
		return TwelveHours, nil
	case time.Hour * 24:
		return OneDay, nil
	case time.Hour * 24 * 7:
		return OneWeek, nil
	case time.Hour * 24 * 7 * 4:
		return OneMonth, nil
	}
	return OneMinute, fmt.Errorf("could not convert duration to resolution: %s", d)
}

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

func (r *CandleResolution) ToDuration() time.Duration {
	switch *r {
	case OneMinute:
		return time.Minute
	case FiveMinutes:
		return time.Minute * 5
	case FifteenMinutes:
		return time.Minute * 15
	case ThirtyMinutes:
		return time.Minute * 30
	case OneHour:
		return time.Hour
	case ThreeHours:
		return time.Hour * 3
	case SixHours:
		return time.Hour * 6
	case TwelveHours:
		return time.Hour * 12
	case OneDay:
		return time.Hour * 24
	case OneWeek:
		return time.Hour * 24 * 7
	case OneMonth:
		return time.Hour * 24 * 7 * 4
	default:
		panic("duration (⊙_⊙)？")
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

type Candles struct {
	Symbol     string
	Resolution CandleResolution
	// first == old
	Candles []*Candle
}

type Candle struct {
	Symbol     string
	Resolution CandleResolution
	// Local date
	Date   time.Time
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume float64
}
