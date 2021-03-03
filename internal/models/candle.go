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

func CandleResolutionFromDuration(d time.Duration) (CandleResolution, error) {
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
	}
	return OneMinute, fmt.Errorf("could not convert duration to resolution: %s", d)
}

func QuoteToModel(q *quote.Quote, symbol string, index int, d time.Duration) *Candle {
	res, err := CandleResolutionFromDuration(d)
	if err != nil {
		res = ""
	}
	return &Candle{
		Symbol:     symbol,
		Resolution: res,
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
	Candles    []*Candle
}

type Candle struct {
	Symbol     string
	Resolution CandleResolution
	Date       time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     float64
}
