package mock

import (
	"DaruBot/internal/cache/candles"
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
	currency = "USDT"

	//market = "coinbase"
	//quoteFrom = quote.NewQuoteFromCoinbase
	market    = "binance-usdt"
	quoteFrom = quote.NewQuoteFromBinance
)

type cacheCandle struct {
	cache *candles.Cache
}

func newCacheCandles(t *testing.T, lg logger.Logger) *cacheCandle {
	candlesCache, err := candles.NewCandleCache("../../../test_data/candles.cache", lg)
	if err != nil {
		t.Fatal(err)
	}

	return &cacheCandle{
		candlesCache,
	}
}

func (c *cacheCandle) Stop() {
	err := c.cache.SaveCache()
	if err != nil {
		panic(err)
	}
}

func newEx(t *testing.T, level logger.Level, from, to time.Time, p *Plutos) (*exchange, func(), error) {
	lg := logger.New(os.Stdout, level)
	ctx, finish := context.WithCancel(context.Background())

	wManager := watcher.NewWatcherManager()

	stand := NewTheWorld(from, to, 1*time.Millisecond)

	cfg := config.GetDefaultConfig()

	downloadCandles = quoteFrom

	cache := newCacheCandles(t, lg)

	if p == nil {
		p = newPlutos(nil, currency)
	}

	mk, err := newExchangeMock(ctx, wManager, lg, cfg, market, cache.cache, stand, p)
	p.SetTickerFunc(mk.getTicker)

	stop := func() {
		finish()
		cache.Stop()
	}

	return mk, stop, err
}

func startWatcher(t *testing.T, mk *exchange) func() {
	ctx, cancel := context.WithCancel(context.Background())

	go func(tt *testing.T) {
		wh := mk.watchers.MustNew("all_events", "", "")

		for {
			select {
			case evt := <-wh.Listen():
				if evt.Is(models.EventError) {
					tt.Fatalf("error: %v", evt.Payload)
				}
				//t.Logf("event type: %v(%v), payload: [%#v] \n", EventToString(evt.Head), evt.Head, evt.Payload)
				fmt.Printf("event: %v(%v), payload: [%#v] \n", evt.GetModuleType(), evt.GetEventName(), evt.Payload)
			case <-ctx.Done():
				mk.watchers.Remove("all_events")
				return
			}
		}
	}(t)

	return cancel
}

func TestCheckPair(t *testing.T) {
	mk, _, err := newEx(t, logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}

	err = mk.CheckSymbol(testPair, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTicker(t *testing.T) {
	mk, stop, err := newEx(t, logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	tk, err := mk.GetTicker(testPair)
	if err != nil {
		t.Fatal(err)
	}

	litter.Dump(tk)
}

func TestGetCandles(t *testing.T) {
	mk, stop, err := newEx(t, logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	cndls, err := mk.GetCandles(testPair, models.OneHour, time.Now().Add(-time.Hour*6), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	litter.Dump(cndls)
}

func TestGetLastCandle(t *testing.T) {
	mk, stop, err := newEx(t, logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	candle, err := mk.GetLastCandle(testPair, models.OneHour)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(candle.Date.String())

	litter.Dump(candle)
}

func TestGetBalance(t *testing.T) {
	w := &models.Wallets{}
	w.WalletType = models.WalletTypeNone
	w.Update(&models.WalletCurrency{
		Name:       currency,
		WalletType: models.WalletTypeNone,
		Balance:    1000,
		Available:  1000,
	})
	w.Update(&models.WalletCurrency{
		Name:       "BTC",
		WalletType: models.WalletTypeNone,
		Balance:    0.003,
		Available:  0.003,
	})

	p := newPlutos(w, currency)

	mk, stop, err := newEx(t, logger.TraceLevel, time.Now().Add(-time.Hour*24*2), time.Now(), p)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	balance, err := mk.GetBalance()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(balance)

	litter.Dump(balance)
}
