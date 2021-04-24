package mock

import (
	"DaruBot/internal/models"
	"math"
	"time"
)

var (
	timeFormat  = "2006-01-02 15:04"
	timeFormatD = "2006-01-02"
)

func getCandle(q *models.Candles, t time.Time) *models.Candle {
	if len(q.Candles) == 0 {
		return nil
	}

	if t.IsZero() {
		return q.Candles[len(q.Candles)-1]
	}

	d := q.Resolution.ToDuration()

	for i := 0; i < len(q.Candles); i++ {
		if q.Candles[i].Date.Sub(t) <= d {
			return q.Candles[i]
		}
	}

	return nil
}

func quoteFormat(t time.Time, format string) string {
	return t.Format(format)
}

func downloadQuote(from, to time.Time, symbol string, resolution models.CandleResolution) (*models.Candles, error) {
	qRes, err := resolution.ToQuoteModel()
	if err != nil {
		return nil, err
	}

	format := timeFormat
	if resolution.ToDuration() >= models.OneDay.ToDuration() {
		format = timeFormatD
	}

	start := quoteFormat(from.UTC(), format)
	end := quoteFormat(to.UTC(), format)

	q, err := downloadCandles(symbol, start, end, qRes)
	if err != nil {
		return nil, err
	}

	cndls := models.QuoteToModels(&q, symbol)

	if len(cndls.Candles) == 1 {
		cndls.Resolution = resolution
		cndls.Candles[0].Resolution = resolution
	}

	return cndls, nil
}

func checkResTiming(d time.Duration, t time.Time) bool {
	switch {
	case math.Mod(float64(t.Unix()), d.Seconds()) == 0:
		return true
	default:
		return false
	}
}
