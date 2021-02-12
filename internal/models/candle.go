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

type Candles struct {
	Pair       string
	Resolution CandleResolution
	Candles    []*Candle
}

type Candle struct {
	Pair       string
	Resolution CandleResolution
	Date       time.Time
	Open       float64
	Close      float64
	High       float64
	Low        float64
	Volume     float64
}
