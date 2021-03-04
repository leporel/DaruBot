/*
Download and collects candles
Downloaded candles will be cached and stored, if requested range are not
represented in cache, download range and combine with existing.
*/
package candles

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/errors"
	"encoding/json"
	"fmt"
	"github.com/alibaba/pouch/pkg/kmutex"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"os"
	"time"
)

var (
	ErrWrongSort          = errors.New("Candles not sorted old > new")
	ErrWrongCandle        = errors.New("Candles are not consistent")
	ErrDownloadLastCandle = errors.New("Last candle not downloaded")
)

type marketCandles []*period

func (m *marketCandles) get(from, to time.Time) (*models.Candles, bool) {
	for _, c := range *m {
		if (c.from.Before(from) || c.from.Equal(from)) &&
			(c.to.Before(to) || c.to.Equal(to)) {
			return c.part(from, to), true
		}
	}
	return nil, false
}

func (m *marketCandles) add(pd *period) {

	var needRefresh bool

	periods := *m

	for i := 0; i < len(periods); i++ {
		needRefresh = canCombine(pd, periods[i])
		if needRefresh {
			break
		}
	}

	periods = append(marketCandles{pd}, periods...)

	if needRefresh {
		old := periods
		refreshed := make(marketCandles, 0)

		for fi := 0; fi < len(old); fi++ {
			tempPeriod := old[fi]
			for i := 1; i < len(old); i++ {
				if canCombine(tempPeriod, old[i]) {
					tempPeriod = combine(tempPeriod, old[i])
					old = append(old[:i], old[i+1:]...)
				}
			}
			refreshed = append(refreshed, tempPeriod)
		}

		periods = refreshed
	}

	*m = periods
}

type period struct {
	to      time.Time
	from    time.Time
	candles *models.Candles
}

func (p *period) part(from, to time.Time) *models.Candles {
	rs := models.Candles{
		Symbol:     p.candles.Symbol,
		Resolution: p.candles.Resolution,
		Candles:    make([]*models.Candle, 0),
	}

	for _, c := range p.candles.Candles {
		if (c.Date.Before(from) || c.Date.Equal(from)) &&
			(c.Date.Before(to) || c.Date.Equal(to)) {
			rs.Candles = append(rs.Candles, c)
		}
	}

	return nil
}

type candlesCache struct {
	cache    *cache.Cache
	filePath string
}

func NewCandleCache(filePath string) (*candlesCache, error) {
	if filePath == "" {
		filePath = "./candles.cache"
	}
	f, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	items := make(map[string]cache.Item)

	if len(raw) != 0 {
		err = json.Unmarshal(raw, &items)
		if err != nil {
			return nil, err
		}
	}

	c := cache.NewFrom(cache.NoExpiration, 10*time.Minute, items)

	rs := &candlesCache{
		cache:    c,
		filePath: filePath,
	}
	return rs, nil
}

func (c *candlesCache) NewMarket(name string, loadFunc loadFunc) *marketCandlesCache {
	return &marketCandlesCache{
		cache:      c.cache,
		market:     name,
		lock:       kmutex.New(),
		loaderFunc: loadFunc,
	}
}

func (c *candlesCache) SaveCache() error {
	f, err := os.OpenFile(c.filePath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	raw, err := json.Marshal(c.cache.Items())
	if err != nil {
		return err
	}

	_, err = f.Write(raw)
	if err != nil {
		return err
	}

	return nil
}

type loadFunc func(from, to time.Time, symbol string, resolution models.CandleResolution) (*models.Candles, error)

type marketCandlesCache struct {
	lock       *kmutex.KMutex
	cache      *cache.Cache
	market     string
	loaderFunc loadFunc
}

func (c *marketCandlesCache) Get(from, to time.Time, symbol string, resolution models.CandleResolution) (*models.Candles, error) {
	var hasUpdate bool
	key := c.makeKey(symbol, resolution)

	c.lock.Lock(key)
	defer c.lock.Unlock(key)

	var lastCandle *models.Candle

	if time.Now().Sub(to) <= resolution.ToDuration() {
		// Last candle requested
		// Always download, because last candle not closed and should not be cached

		candles, err := c.loaderFunc(to.Add(-resolution.ToDuration()), to, symbol, resolution)
		if err != nil {
			return nil, err
		}

		if len(candles.Candles) == 0 {
			return nil, ErrDownloadLastCandle
		}

		lastCandle = candles.Candles[len(candles.Candles)-1]

		to = to.Add(-resolution.ToDuration())
	}

	// Get resolutions from stored for symbol
	loaded, found := c.get(key)
	if !found {
		candles, err := c.loaderFunc(from, to, symbol, resolution)
		if err != nil {
			return nil, err
		}

		loaded = marketCandles{&period{
			to:      candles.Candles[len(candles.Candles)-1].Date,
			from:    candles.Candles[0].Date,
			candles: candles,
		}}

		hasUpdate = true
	}
	periods := loaded.(*marketCandles)

	// Get range of candles
	candles, exist := periods.get(from, to)
	if !exist {
		fetched, err := c.loaderFunc(from, to, symbol, resolution)
		if err != nil {
			return nil, err
		}

		candles = fetched
		periods.add(&period{
			to:      candles.Candles[len(candles.Candles)-1].Date,
			from:    candles.Candles[0].Date,
			candles: candles,
		})

		hasUpdate = true
	}

	if err := VerifyCandles(candles); err != nil {
		return nil, err
	}

	if hasUpdate {
		c.set(key, *periods)
	}

	if lastCandle != nil {
		candles.Candles = append(candles.Candles, lastCandle)
	}

	return candles, nil
}

func (c *marketCandlesCache) makeKey(symbol string, resolution models.CandleResolution) string {
	return fmt.Sprintf("%s_%s_%s", c.market, resolution, symbol)
}

func (c *marketCandlesCache) get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *marketCandlesCache) set(key string, m marketCandles) {
	c.cache.Set(key, &m, cache.NoExpiration)
}

func canCombine(period1, period2 *period) bool {
	if (period1.to.After(period2.from) && period1.to.Equal(period2.from)) ||
		(period1.to.Before(period2.to) && period1.to.Equal(period2.to)) {
		return true
	}
	if (period1.to.After(period2.from) && period1.to.Equal(period2.from)) ||
		(period1.from.Before(period2.to) && period1.from.Equal(period2.to)) {
		return true
	}
	return false
}

func combine(period1, period2 *period) *period {
	rs := &period{
		to:   time.Time{},
		from: time.Time{},
		candles: &models.Candles{
			Symbol:     period1.candles.Symbol,
			Resolution: period1.candles.Resolution,
			Candles:    make([]*models.Candle, 0),
		},
	}

	first, second := period1, period2
	if period1.from.After(period2.from) {
		first, second = period2, period1
	}

	var last time.Time
	for _, c := range first.candles.Candles {
		rs.candles.Candles = append(rs.candles.Candles, c)
		last = c.Date
	}

	for _, c := range second.candles.Candles {
		if c.Date.After(last) {
			rs.candles.Candles = append(rs.candles.Candles, c)
		}
	}

	rs.from = first.from
	rs.to = second.to
	if second.to.Before(first.to) || second.to.Equal(first.to) {
		rs.to = first.to
	}

	return rs
}

func VerifyCandles(candles *models.Candles) error {
	if len(candles.Candles) > 1 {
		for i := 1; i < len(candles.Candles); i++ {
			if candles.Candles[i-1].Date.After(candles.Candles[i].Date) {
				return ErrWrongSort
			}
			if candles.Candles[i-1].Date.Sub(candles.Candles[i].Date) != candles.Resolution.ToDuration() {
				return ErrWrongCandle
			}
		}
	}
	return nil
}
