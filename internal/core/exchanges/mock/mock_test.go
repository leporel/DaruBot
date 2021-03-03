package mock

import (
	"DaruBot/internal/config"
	"DaruBot/internal/models"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/watcher"
	"context"
	"fmt"
	"github.com/markcheno/go-quote"
	"github.com/sanity-io/litter"
	"os"
	"testing"
	"time"
)

var (
	//testPair = "BTC-USD" // coinbase
	testPair = "BTCUSDT" // binance-usdt

	//cryptoMarket = "coinbase"
	//quoteFrom = quote.NewQuoteFromCoinbase
	market    = "binance-usdt"
	quoteFrom = quote.NewQuoteFromBinance
)

func newEx(level logger.Level, from, to time.Time) (*exchange, func(), error) {
	lg := logger.New(os.Stdout, level)
	ctx, finish := context.WithCancel(context.Background())

	wManager := watcher.NewWatcherManager()

	stand := NewTheWorld(from, to, 1*time.Millisecond)

	cfg := config.GetDefaultConfig()

	mk, err := newExchangeMock(ctx, wManager, lg, cfg, market, quoteFrom, stand)

	return mk, finish, err
}

func startWatcher(t *testing.T, mk *exchange) func() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		wh := mk.watchers.MustNew("all_events", "", "")

		for {
			select {
			case evt := <-wh.Listen():
				if evt.Is(models.EventError) {
					t.Fatalf("error: %v", evt.Payload)
				}
				//t.Logf("event type: %v(%v), payload: [%#v] \n", EventToString(evt.Head), evt.Head, evt.Payload)
				fmt.Printf("event: %v(%v), payload: [%#v] \n", evt.GetModuleType(), evt.GetEventName(), evt.Payload)
			case <-ctx.Done():
				mk.watchers.Remove("all_events")
				return
			}
		}
	}()

	return cancel
}

func TestCheckPair(t *testing.T) {
	mk, _, err := newEx(logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	err = mk.CheckSymbol(testPair, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTicker(t *testing.T) {
	mk, _, err := newEx(logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	tk, err := mk.GetTicker(testPair)
	if err != nil {
		t.Fatal(err)
	}

	litter.Dump(tk)
}

func TestGetCandles(t *testing.T) {
	mk, _, err := newEx(logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	candles, err := mk.GetCandles(testPair, models.OneHour, time.Now().Add(-time.Hour*6), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	litter.Dump(candles)
}

func TestGetLastCandle(t *testing.T) {
	mk, _, err := newEx(logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	candle, err := mk.GetLastCandle(testPair, models.OneHour)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(candle.Date.String())

	litter.Dump(candle)
}
