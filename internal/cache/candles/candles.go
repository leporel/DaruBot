/*
Download and collects candles
Downloaded candles will be cached and stored, if requested range are not
represented in cache, download range and combine with existing.
*/
package candles

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/errors"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/tools/numbers"
	"encoding/json"
	"fmt"
	"github.com/alibaba/pouch/pkg/kmutex"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"os"
	"time"
)

var (
	version = "v1"

	ErrWrongSort          = errors.New("Candles not sorted old > new")
	ErrWrongCandle        = errors.New("Candles are not consistent")
	ErrDownloadLastCandle = errors.New("Last candle not downloaded")
)

type period struct {
	From    time.Time
	To      time.Time
	Candles *models.Candles
}

type Cache struct {
	cache    *cache.Cache
	filePath string
	lg       logger.Logger
}

type marketCandles struct {
	Periods []*period
}

type MarketCandlesCache struct {
	lock       *kmutex.KMutex
	cache      *cache.Cache
	market     string
	loaderFunc loadFunc
	lg         logger.Logger
}

type cacheCollection struct {
	Version string
	Items   map[string]item
}

type item struct {
	Object     *marketCandles
	Expiration int64
}

type loadFunc func(from, to time.Time, symbol string, resolution models.CandleResolution) (*models.Candles, error)

func NewCandleCache(filePath string, lg logger.Logger) (*Cache, error) {
	if filePath == "" {
		return nil, fmt.Errorf("filePath empty")
	}
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	data := cacheCollection{
		Version: version,
		Items:   nil,
	}

	data.Items = make(map[string]item)

	if len(raw) != 0 {
		err = json.Unmarshal(raw, &data)
		if err != nil {
			return nil, err
		}
	}

	cItems := make(map[string]cache.Item, 0)

	if data.Version == version {
		cItems = make(map[string]cache.Item, len(data.Items))

		for s, item := range data.Items {
			cItems[s] = cache.Item{
				Object:     item.Object,
				Expiration: item.Expiration,
			}
		}
	}

	log := lg.WithPrefix("module", "candles cache")
	c := cache.NewFrom(cache.NoExpiration, 10*time.Minute, cItems)

	rs := &Cache{
		cache:    c,
		filePath: filePath,
		lg:       log,
	}

	log.Debug("cache loaded", filePath)

	return rs, nil
}

func (c *Cache) SaveCache() error {
	f, err := os.OpenFile(c.filePath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	items := c.cache.Items()

	data := cacheCollection{
		Version: version,
		Items:   nil,
	}

	data.Items = make(map[string]item, len(items))

	for s, it := range items {
		data.Items[s] = item{
			Object:     it.Object.(*marketCandles),
			Expiration: it.Expiration,
		}
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = f.Write(raw)
	if err != nil {
		return err
	}

	c.lg.Debug("cache saved")

	return nil
}

func (c *Cache) GetMarket(name string, loadFunc loadFunc) *MarketCandlesCache {
	c.lg.Debug("cache get market", name)
	return &MarketCandlesCache{
		cache:      c.cache,
		market:     name,
		lock:       kmutex.New(),
		loaderFunc: loadFunc,
		lg:         c.lg.WithPrefix("market", name),
	}
}

func (c *MarketCandlesCache) Get(from, to time.Time, symbol string, resolution models.CandleResolution) (*models.Candles, error) {
	c.lg.Tracef("get candles [%s-%s] %s - %s", c.market, symbol, from.Format(time.RFC822Z), to.Format(time.RFC822Z))
	var rs *models.Candles
	var hasUpdate bool
	key := c.makeKey(symbol, resolution)
	c.lock.Lock(key)
	defer c.lock.Unlock(key)

	start, end, err := normalizeDate(from, to, resolution)
	if err != nil {
		return nil, err
	}

	var lastCandle *models.Candle

	if time.Now().Sub(end) <= 0 {
		// Last candle requested
		// Always download, because last candle not closed and should not be cached

		c.lg.Trace("download last candle")

		candles, err := c.loaderFunc(time.Now().Add(-resolution.ToDuration()), time.Now(), symbol, resolution)
		if err != nil {
			return nil, err
		}

		if len(candles.Candles) == 0 {
			return nil, ErrDownloadLastCandle
		}

		lastCandle = candles.Candles[len(candles.Candles)-1]

		end = end.Add(-resolution.ToDuration())
	}

	// Get resolutions from stored for symbol
	loaded, found := c.load(key)
	if !found {
		c.lg.Trace("cached candles for this market not found, make new...")

		loaded = &marketCandles{
			Periods: []*period{},
		}
	} else {
		c.lg.Trace("cached candles for this market found")
	}
	collection := loaded.(*marketCandles)

	// Get range of candles
	candles, exist := collection.get(start, end)
	if !exist {
		// Get extra old candle to cache
		startEx := start.Add(-resolution.ToDuration())

		c.lg.Tracef("cached candles for %s - %s periods not found, download...", startEx.Format(time.RFC822Z), end.Format(time.RFC822Z))
		fetched, err := c.loaderFunc(startEx, end, symbol, resolution)
		if err != nil {
			return nil, err
		}

		candles = fetched
		pd := &period{
			To:      candles.Candles[len(candles.Candles)-1].Date,
			From:    candles.Candles[0].Date,
			Candles: candles,
		}

		collection.add(pd)

		rs = pd.part(start, end)

		if len(pd.Candles.Candles) > 0 {
			hasUpdate = true
		}
	} else {
		rs = candles
		c.lg.Trace("cached candles for this periods found")
	}

	if lastCandle != nil {
		c.lg.Trace("append last candle")
		rs.Candles = append(rs.Candles, lastCandle)
	}

	if err = c.VerifyCandles(rs); err != nil {
		return nil, err
	}

	if hasUpdate {
		c.lg.Trace("update cache")
		c.set(key, collection)
	}

	return rs, nil
}

func (c *MarketCandlesCache) makeKey(symbol string, resolution models.CandleResolution) string {
	return fmt.Sprintf("%s_%s_%s", c.market, resolution, symbol)
}

func (c *MarketCandlesCache) load(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *MarketCandlesCache) set(key string, m *marketCandles) {
	c.cache.Set(key, m, cache.NoExpiration)
}

func (p *period) part(from, to time.Time) *models.Candles {
	rs := &models.Candles{
		Symbol:     p.Candles.Symbol,
		Resolution: p.Candles.Resolution,
		Candles:    make([]*models.Candle, 0),
	}

	for _, c := range p.Candles.Candles {
		if (c.Date.After(from) || c.Date.Equal(from)) &&
			(c.Date.Before(to) || c.Date.Equal(to)) {
			rs.Candles = append(rs.Candles, c)
		}
	}

	return rs
}

func (m *marketCandles) get(from, to time.Time) (*models.Candles, bool) {
	for _, c := range m.Periods {
		if (c.From.Before(from) || c.From.Equal(from)) &&
			(c.To.After(to) || c.To.Equal(to)) {
			return c.part(from, to), true
		}
	}
	return nil, false
}

func (m *marketCandles) add(pd *period) {

	var needRefresh bool

	periods := m.Periods

	for i := 0; i < len(periods); i++ {
		needRefresh = canCombine(pd, periods[i])
		if needRefresh {
			break
		}
	}

	periods = append([]*period{pd}, periods...)

	for needRefresh {
		needRefresh = false
		old := periods
		refreshed := make([]*period, 0, len(old))

		for fi := 0; fi < len(old); fi++ {
			tempPeriod := old[fi]
			for i := 1; i < len(old); i++ {
				if canCombine(tempPeriod, old[i]) {
					tempPeriod = combine(tempPeriod, old[i])
					old = append(old[:i], old[i+1:]...)
					needRefresh = true
				}
			}
			refreshed = append(refreshed, tempPeriod)
		}

		periods = refreshed
	}

	m.Periods = periods
}

func canCombine(period1, period2 *period) bool {
	if (period1.To.After(period2.From) || period1.To.Equal(period2.From)) &&
		(period1.To.Before(period2.To) || period1.To.Equal(period2.To)) {
		return true
	}
	if (period1.To.After(period2.From) || period1.To.Equal(period2.From)) &&
		(period1.From.Before(period2.To) || period1.From.Equal(period2.To)) {
		return true
	}
	return false
}

func combine(period1, period2 *period) *period {
	rs := &period{
		To:   time.Time{},
		From: time.Time{},
		Candles: &models.Candles{
			Symbol:     period1.Candles.Symbol,
			Resolution: period1.Candles.Resolution,
			Candles:    make([]*models.Candle, 0),
		},
	}

	first, second := period1, period2
	if period1.From.After(period2.From) {
		first, second = period2, period1
	}

	var last time.Time
	for _, c := range first.Candles.Candles {
		rs.Candles.Candles = append(rs.Candles.Candles, c)
		last = c.Date
	}

	for _, c := range second.Candles.Candles {
		if c.Date.After(last) {
			rs.Candles.Candles = append(rs.Candles.Candles, c)
		}
	}

	rs.From = first.From
	rs.To = second.To
	if second.To.Before(first.To) || second.To.Equal(first.To) {
		rs.To = first.To
	}

	return rs
}

func (c *MarketCandlesCache) VerifyCandles(candles *models.Candles) error {
	if len(candles.Candles) > 1 {
		for i := 1; i < len(candles.Candles); i++ {
			if candles.Candles[i-1].Date.After(candles.Candles[i].Date) {
				return ErrWrongSort
			}
			if candles.Candles[i-1].Date.Location() != candles.Candles[i].Date.Location() {
				c.lg.Warn("time have different locations, probably candles date have changed")
			}
			if candles.Candles[i-1].Symbol != candles.Candles[i].Symbol {
				c.lg.Tracef("candle check consistent failed \ncandle1: %s \ncandle2: %s\n",
					candles.Candles[i-1].Symbol,
					candles.Candles[i].Symbol)
				return ErrWrongCandle
			}
			if candles.Candles[i].Date.Sub(candles.Candles[i-1].Date) != candles.Resolution.ToDuration() {
				c.lg.Tracef("candle check consistent failed \ncandle1: %s \ncandle2: %s \ngot: %v \nwant: %v\n",
					candles.Candles[i-1].Date, candles.Candles[i].Date,
					candles.Candles[i].Date.Sub(candles.Candles[i-1].Date),
					candles.Resolution.ToDuration())
				return ErrWrongCandle
			}
		}
	}
	return nil
}

func normalizeDate(from, to time.Time, resolution models.CandleResolution) (time.Time, time.Time, error) {
	var start, end time.Time

	from = from.UTC()
	to = to.UTC()

	switch resolution {
	case models.OneMinute:
		end = time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), to.Minute(), 59, 0, time.UTC)

		start = time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), from.Minute(), 0, 0, time.UTC)
	case models.FiveMinutes:
		minute := numbers.NumberRoundTo(to.Minute(), 5)
		end = time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), minute, 59, 0, time.UTC)

		minute = numbers.NumberRoundTo(from.Minute(), -5)
		start = time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), minute, 0, 0, time.UTC)
	case models.FifteenMinutes:
		minute := numbers.NumberRoundTo(to.Minute(), 15)
		end = time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), minute, 59, 0, time.UTC)

		minute = numbers.NumberRoundTo(from.Minute(), -15)
		start = time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), minute, 0, 0, time.UTC)
	case models.ThirtyMinutes:
		minute := numbers.NumberRoundTo(to.Minute(), 30)
		end = time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), minute, 59, 0, time.UTC)

		minute = numbers.NumberRoundTo(from.Minute(), -30)
		start = time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), minute, 0, 0, time.UTC)
	case models.OneHour:
		end = time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), 59, 59, 0, time.UTC)

		start = time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), 0, 0, 0, time.UTC)
	case models.ThreeHours:
		hour := numbers.NumberRoundTo(to.Hour(), 3)
		end = time.Date(to.Year(), to.Month(), to.Day(), hour, 59, 59, 0, time.UTC)

		hour = numbers.NumberRoundTo(from.Hour(), -3)
		start = time.Date(from.Year(), from.Month(), from.Day(), hour, 0, 0, 0, time.UTC)
	case models.SixHours:
		hour := numbers.NumberRoundTo(to.Hour(), 6)
		end = time.Date(to.Year(), to.Month(), to.Day(), hour, 59, 59, 0, time.UTC)

		hour = numbers.NumberRoundTo(from.Hour(), -6)
		start = time.Date(from.Year(), from.Month(), from.Day(), hour, 0, 0, 0, time.UTC)
	case models.TwelveHours:
		hour := numbers.NumberRoundTo(to.Hour(), 12)
		end = time.Date(to.Year(), to.Month(), to.Day(), hour, 59, 59, 0, time.UTC)

		hour = numbers.NumberRoundTo(from.Hour(), -12)
		start = time.Date(from.Year(), from.Month(), from.Day(), hour, 0, 0, 0, time.UTC)
	case models.OneDay:
		end = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, time.UTC)

		start = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	case models.OneWeek:
		day := numbers.NumberRoundTo(to.Day(), 7)
		end = time.Date(to.Year(), to.Month(), day, 23, 59, 59, 0, time.UTC)

		day = numbers.NumberRoundTo(from.Day(), -7)
		start = time.Date(from.Year(), from.Month(), day, 0, 0, 0, 0, time.UTC)
	case models.OneMonth:
		month := to.Month()
		if to.Day() != 1 {
			month = month + 1
		}
		end = time.Date(to.Year(), month, 1, 0, 0, 0, 0, time.UTC)

		start = time.Date(from.Year(), month, 1, 0, 0, 0, 0, time.UTC)
	default:
		return from, to, fmt.Errorf("unknown resolution")
	}

	return start, end, nil
}
