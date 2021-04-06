package models

import (
	"fmt"
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

func (r CandleResolution) String() string {
	return string(r)
}

func (r CandleResolution) ToDuration() time.Duration {
	switch r {
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
