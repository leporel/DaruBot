package candles

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/logger"
	"fmt"
	"github.com/markcheno/go-quote"
	"math"
	"os"
	"testing"
	"time"
)

var (
	//testPair = "BTC-USD" // coinbase
	testPair = "BTCUSDT" // binance-usdt

	//market = "coinbase"
	//quoteFrom = quote.NewQuoteFromCoinbase
	market    = "binance"
	quoteFrom = quote.NewQuoteFromBinance

	timeFormat  = "2006-01-02 15:04"
	timeFormatD = "2006-01-02"
)

func quoteFormat(t time.Time, format string) string {
	return t.Format(format)
}

func downloadQuote(from, to time.Time, symbol string, resolution models.CandleResolution) (*models.Candles, error) {
	qRes, err := resolution.ToQuoteModel()
	if err != nil {
		return nil, err
	}

	//from = from.UTC()
	//to = to.UTC()

	format := timeFormat
	if resolution.ToDuration() >= models.OneDay.ToDuration() {
		format = timeFormatD
	}

	start := quoteFormat(from, format)
	end := quoteFormat(to, format)

	fmt.Printf("downlaod from %s to %s (%#v)\n", start, end, from.Location().String())
	q, err := quoteFrom(symbol, start, end, qRes)
	if err != nil {
		return nil, err
	}

	return models.QuoteToModels(&q, symbol), nil
}

func newCache(t *testing.T) *candlesCache {
	writer := os.Stdout
	lg := logger.New(writer, logger.TraceLevel)

	storageCache, err := NewCandleCache("../../../test_data/candles.cache", lg)
	if err != nil {
		t.Fatal(err)
	}

	return storageCache
}

func TestPeriods(t *testing.T) {
	from := time.Date(2020, time.November, 27, 0, 0, 0, 0, time.Local)
	to := time.Date(2020, time.December, 1, 0, 0, 0, 0, time.Local)
	res := models.OneDay

	candles, err := downloadQuote(from, to, testPair, res)
	if err != nil {
		t.Fatal(err)
	}
	collection := &marketCandles{
		Periods: []*period{{
			To:      candles.Candles[len(candles.Candles)-1].Date,
			From:    candles.Candles[0].Date,
			Candles: candles,
		}},
	}

	for _, p := range collection.Periods {
		t.Logf("%s -> %s\n", p.From.Format(time.Stamp), p.To.Format(time.Stamp))
		for _, cd := range p.Candles.Candles {
			t.Log("\t", cd.Date.Format(time.Stamp), cd.Open, cd.Close)
		}
	}

	collection.add(collection.Periods[0])

	if len(collection.Periods) != 1 {
		t.Fatalf("wrong len periods, wanted: %v, got: %v", 1, len(collection.Periods))
	}

	from2 := time.Date(2020, time.December, 2, 0, 0, 0, 0, time.Local)
	to2 := time.Date(2020, time.December, 3, 0, 0, 0, 0, time.Local)

	candles2, err := downloadQuote(from2, to2, testPair, res)
	if err != nil {
		t.Fatal(err)
	}
	collection.add(&period{
		To:      candles2.Candles[len(candles2.Candles)-1].Date,
		From:    candles2.Candles[0].Date,
		Candles: candles2,
	})

	for _, p := range collection.Periods {
		t.Logf("%s -> %s\n", p.From.Format(time.Stamp), p.To.Format(time.Stamp))
		for _, cd := range p.Candles.Candles {
			t.Log("\t", cd.Date.Format(time.Stamp), cd.Open, cd.Close)
		}
	}

	if len(collection.Periods) != 2 {
		t.Fatal("periods are not stacked")
	}

	from3 := time.Date(2020, time.November, 29, 0, 0, 0, 0, time.Local)
	to3 := time.Date(2020, time.December, 2, 0, 0, 0, 0, time.Local)

	candles3, err := downloadQuote(from3, to3, testPair, res)
	if err != nil {
		t.Fatal(err)
	}
	collection.add(&period{
		To:      candles3.Candles[len(candles3.Candles)-1].Date,
		From:    candles3.Candles[0].Date,
		Candles: candles3,
	})

	for _, p := range collection.Periods {
		t.Logf("%s -> %s\n", p.From.Format(time.Stamp), p.To.Format(time.Stamp))
		for _, cd := range p.Candles.Candles {
			t.Log("\t", cd.Date.Format(time.Stamp), cd.Open, cd.Close)
		}
	}

	if len(collection.Periods) != 1 {
		t.Fatal("periods are not combined")
	}

	candles4, err := downloadQuote(from, to2, testPair, res)
	if err != nil {
		t.Fatal(err)
	}
	collection.add(&period{
		To:      candles4.Candles[len(candles4.Candles)-1].Date,
		From:    candles4.Candles[0].Date,
		Candles: candles4,
	})

	for _, p := range collection.Periods {
		t.Logf("%s -> %s\n", p.From.Format(time.Stamp), p.To.Format(time.Stamp))
		for _, cd := range p.Candles.Candles {
			t.Log("\t", cd.Date.Format(time.Stamp), cd.Open, cd.Close)
		}
	}

	if len(collection.Periods) != 1 {
		t.Fatal("periods are not combined")
	}

}

func TestLoadCache(t *testing.T) {
	err := os.Remove("../../../test_data/candles.cache")
	switch err.(type) {
	case nil:
	case *os.PathError:
	default:
		t.Fatal(err)
	}
	newCache(t)
}

func TestSaveLoad(t *testing.T) {
	err := os.Remove("../../../test_data/candles.cache")
	switch err.(type) {
	case nil:
	case *os.PathError:
	default:
		t.Fatal(err)
	}

	from := time.Date(2020, time.November, 27, 0, 0, 0, 0, time.Local)
	to := time.Date(2020, time.December, 1, 5, 0, 0, 0, time.Local)

	res := models.OneDay
	marketName := fmt.Sprint(market, time.Now().Unix())
	start, end, _ := normalizeDate(from, to, res)
	days := int(math.Ceil(end.Sub(start).Hours() / 24))

	c := newCache(t)
	mk := c.GetMarket(marketName, downloadQuote)

	candles, err := mk.Get(from, to, testPair, res)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("from %s to %s", from, to)
	for _, cd := range candles.Candles {
		t.Log("\t", cd.Date.Format(time.Stamp), cd.Open, cd.Close)
	}

	if len(candles.Candles) != days {
		t.Fatalf("wrong candles len, wanted: %v, got: %v", days, len(candles.Candles))
	}
	if candles.Candles[0].Resolution != res {
		t.Fatalf("wrong candles len, wanted: %v, got: %v", res, candles.Candles[0].Resolution)
	}

	err = c.SaveCache()
	if err != nil {
		t.Fatal(err)
	}

	c2 := newCache(t)
	mk2 := c2.GetMarket(marketName, downloadQuote)

	key := mk2.makeKey(testPair, res)

	loaded, found := mk2.load(key)
	if !found {
		t.Fatal("cache not found")
	}
	collection := loaded.(*marketCandles)

	for _, p := range collection.Periods {
		t.Logf("%s -> %s\n", p.From.Format(time.Stamp), p.To.Format(time.Stamp))
		for _, cd := range p.Candles.Candles {
			t.Log("\t", cd.Date.Format(time.Stamp), cd.Open, cd.Close)
		}
	}

	t.Logf("\nGet part from: %s to: %s\n", start, end)

	candles, exist := collection.get(start, end)
	if !exist {
		t.Fatal("periods not exist")
	}

	for _, cd := range candles.Candles {
		t.Log(cd.Date.Format(time.Stamp), cd.Open, cd.Close)
	}

	if len(candles.Candles) != days {
		t.Fatalf("wrong candles len, wanted: %v, got: %v", days, len(candles.Candles))
	}
	if candles.Candles[0].Resolution != res {
		t.Fatalf("wrong candles len, wanted: %v, got: %v", res, candles.Candles[0].Resolution)
	}

}

func TestGetWithLastCandle(t *testing.T) {
	err := os.Remove("../../../test_data/candles.cache")
	switch err.(type) {
	case nil:
	case *os.PathError:
	default:
		t.Fatal(err)
	}

	from := time.Now().AddDate(0, 0, -3)
	//to := time.Date(2021, time.April, 6, 2, 59, 59, 0, time.Local)
	to := time.Now()

	res := models.OneDay

	marketName := fmt.Sprint(market, time.Now().Unix())
	start, end, _ := normalizeDate(from, to, res)
	days := int(math.Ceil(end.Sub(start).Hours() / 24))

	//t.Log(to)
	//t.Log(end)
	//t.Log(time.Now().Sub(end), res.ToDuration(), time.Now().Sub(end) <= res.ToDuration())

	c := newCache(t)
	mk := c.GetMarket(marketName, downloadQuote)

	candles, err := mk.Get(from, to, testPair, res)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("from %s to %s", from, to)
	for _, cd := range candles.Candles {
		t.Log("\t", cd.Date.Format(time.Stamp), cd.Open, cd.Close)
	}

	if len(candles.Candles) != days {
		t.Fatalf("wrong candles len, wanted: %v, got: %v", days, len(candles.Candles))
	}
	if candles.Candles[0].Resolution != res {
		t.Fatalf("wrong candles len, wanted: %v, got: %v", res, candles.Candles[0].Resolution)
	}

	key := mk.makeKey(testPair, res)

	loaded, found := mk.load(key)
	if !found {
		t.Fatal("cache not found")
	}
	collection := loaded.(*marketCandles)

	t.Log("\nCache:")
	for _, p := range collection.Periods {
		t.Logf("%s -> %s\n", p.From.Format(time.Stamp), p.To.Format(time.Stamp))
		for _, cd := range p.Candles.Candles {
			t.Log("\t", cd.Date.Format(time.Stamp), cd.Open, cd.Close)
		}
	}

}
