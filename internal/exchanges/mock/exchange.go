package mock

import (
	"DaruBot/internal/cache/candles"
	"DaruBot/internal/config"
	exchanges2 "DaruBot/internal/exchanges"
	"DaruBot/internal/models"
	"DaruBot/internal/models/exchanges"
	"DaruBot/pkg/errors"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/tools/numbers"
	"DaruBot/pkg/watcher"
	"context"
	"github.com/markcheno/go-quote"
	"time"
)

var (
	downloadCandles quoteFunc

	supportEvents = watcher.EventsMap{
		//models.EventError,

		//models.EventTickerState,
		//models.EventCandleState,

		//models.EventOrderNew,
		//models.EventOrderFilled,
		//models.EventOrderCancel,
		//models.EventOrderPartiallyFilled,
		//models.EventOrderUpdate,

		//models.EventPositionNew,
		//models.EventPositionClosed,
		//models.EventPositionUpdate,

		//models.EventWalletUpdate,
	}
)

type (
	quoteFunc = func(symbol string, startDate, endDate string, period quote.Period) (quote.Quote, error)
)

type exchange struct {
	market    string
	quoteFunc quoteFunc
	dio       *TheWorld

	ctx context.Context
	log logger.Logger
	cfg config.Configurations

	ready     bool
	readyChan chan interface{}

	watchers     *watcher.Manager
	cacheCandles *candles.MarketCandlesCache

	lastUpdate    time.Time
	subscriptions models.Subscriptions
}

func NewExchangeMock(ctx context.Context,
	wManager *watcher.Manager,
	lg logger.Logger,
	cfg config.Configurations,
	market string, candlesCache *candles.Cache, quoteF quoteFunc,
	stand *TheWorld) (exchanges2.CryptoExchange, error) {
	downloadCandles = quoteF
	return newExchangeMock(ctx, wManager, lg, cfg, market, candlesCache, stand)
}

func newExchangeMock(ctx context.Context,
	wManager *watcher.Manager,
	lg logger.Logger,
	cfg config.Configurations,
	market string, candlesCache *candles.Cache,
	stand *TheWorld) (*exchange, error) {

	if !quote.ValidMarket(market) {
		return nil, errors.New("market not supported")
	}

	err := wManager.RegisterEvents(exchanges.ExchangeTypeMock.String(), supportEvents)
	if err != nil {
		return nil, err
	}

	mc := candlesCache.GetMarket(market, downloadQuote)

	rs := &exchange{
		market:       market,
		lastUpdate:   time.Time{},
		ctx:          ctx,
		log:          lg.WithPrefix("exchange", "Mock"),
		cfg:          cfg,
		ready:        false,
		readyChan:    make(chan interface{}, 1),
		watchers:     wManager,
		cacheCandles: mc,
		dio:          stand,
	}

	return rs, nil
}

func (e *exchange) Connect() error {
	e.ready = true
	close(e.readyChan)

	e.work()

	return nil
}

func (e *exchange) work() {
	e.dio.Run()

}

func (e *exchange) Disconnect() {
	e.readyChan = make(chan interface{}, 1)
	e.ready = false
}

func (e *exchange) IsReady() bool {
	return e.ready
}

func (e *exchange) Ready() <-chan interface{} {
	return e.readyChan
}

func (e *exchange) SupportEvents() watcher.EventsMap {
	return supportEvents
}

// GetTicker download quote and return emulated ticker
func (e *exchange) GetTicker(symbol string) (*models.Ticker, error) {
	e.dio.TimeStop()
	defer e.dio.TimeStart()
	curTime := e.dio.CurrentTime()

	e.log.Tracef("get ticker, time: %s (UTC %s)", quoteFormat(curTime, timeFormat), quoteFormat(curTime.UTC(), timeFormat))

	candle, err := e.getCandle(symbol, models.OneDay, curTime)
	if err != nil {
		return nil, err
	}

	//e.log.Tracef("daily candle: %+v", candle)

	dayState := models.TickerState{
		High:    candle.High,
		Low:     candle.Low,
		Volume:  candle.Volume,
		BID:     500,
		BIDSize: candle.Volume / 2,
		ASK:     500,
		ASKSize: candle.Volume / 2,
	}

	candle, err = e.getCandle(symbol, models.OneMinute, curTime)
	if err != nil {
		return nil, err
	}

	//e.log.Tracef("minute candle: %+v", candle)

	ticker := &models.Ticker{
		Symbol:   symbol,
		Price:    numbers.GetRandFloat(candle.Low, candle.High),
		Exchange: exchanges.ExchangeTypeMock,
		State:    dayState,
	}

	e.log.Tracef("formed ticker: %+v", ticker)

	return ticker, nil
}

// getCandle function to help GetTicker to emulate the current ticker
func (e *exchange) getCandle(symbol string, res models.CandleResolution, dioTime time.Time) (*models.Candle, error) {
	var from, to time.Time

	switch res {
	case models.OneDay:
		from = time.Date(dioTime.Year(), dioTime.Month(), dioTime.Day()-5, 0, 0, 0, 0, time.Local)
		to = time.Date(dioTime.Year(), dioTime.Month(), dioTime.Day(), 23, 59, 59, 0, time.Local)
	case models.OneMinute:
		from = time.Date(dioTime.Year(), dioTime.Month(), dioTime.Day(), dioTime.Hour(), dioTime.Minute()-5, 0, 0, time.Local)
		to = time.Date(dioTime.Year(), dioTime.Month(), dioTime.Day(), dioTime.Hour(), dioTime.Minute()+5, 59, 0, time.Local)
	default:
		return nil, errors.New("period not set")
	}

	cndls, err := e.cacheCandles.Get(from, to, symbol, res)
	if err != nil {
		return nil, err
	}

	cndl := getCandle(cndls, dioTime)

	return cndl, nil
}

func (e *exchange) GetCandles(symbol string, resolution models.CandleResolution, from time.Time, to time.Time) (*models.Candles, error) {
	e.dio.TimeStop()
	defer e.dio.TimeStart()

	e.log.Tracef("get candles, from %s to %s", from, to)

	if !from.IsZero() && to.After(from) {
		cndls, err := e.cacheCandles.Get(from, to, symbol, resolution)
		if err != nil {
			return nil, err
		}
		return cndls, nil
	}

	return nil, exchanges2.ErrInvalidRequestParams
}

func (e *exchange) GetLastCandle(symbol string, resolution models.CandleResolution) (*models.Candle, error) {
	e.dio.TimeStop()
	defer e.dio.TimeStart()

	start := e.dio.CurrentTime().Add(-resolution.ToDuration())
	end := e.dio.CurrentTime()

	//e.log.Tracef("get last candle, from %s", end)

	cndls, err := e.cacheCandles.Get(start, end, symbol, resolution)
	if err != nil {
		return nil, err
	}

	cnld := cndls.Candles[len(cndls.Candles)-1]

	e.log.Tracef("get last candle, time %s \n last candle: %s", end.Format(time.RFC822Z), cnld.Date.Format(time.RFC822Z))

	return cnld, nil
}

func (e *exchange) GetSubscriptions() *models.Subscriptions {
	panic("implement me")
}

func (e *exchange) SubscribeTicker(symbol string) (subID string, err error) {
	panic("implement me")
}

func (e *exchange) SubscribeCandles(symbol string, resolution models.CandleResolution) (subID string, err error) {
	panic("implement me")
}

func (e *exchange) Unsubscribe(subID string) error {
	panic("implement me")
}

func (e *exchange) CheckSymbol(symbol string, margin bool) error {
	list, err := quote.NewMarketList(e.market)
	if err != nil {
		return err
	}
	for _, i2 := range list {
		if i2 == symbol {
			return nil
		}
	}
	return exchanges2.ErrSymbolNotSupported
}

func (e *exchange) GetOrders() ([]*models.Order, error) {
	panic("implement me")
}

func (e *exchange) GetPositions() ([]*models.Position, error) {
	panic("implement me")
}

func (e *exchange) GetWallets() ([]*models.Wallets, error) {
	panic("implement me")
}

func (e *exchange) GetBalance() (models.BalanceUSD, error) {
	panic("implement me")
}

func (e *exchange) HasUpdates(t time.Time) bool {
	return t.Before(e.lastUpdate)
}

func (e *exchange) PutOrder(order *models.PutOrder) (*models.Order, error) {
	panic("implement me")
}

func (e *exchange) UpdateOrder(orderID string, price float64, priceStop float64, amount float64) (*models.Order, error) {
	panic("implement me")
}

func (e *exchange) CancelOrder(order *models.Order) error {
	panic("implement me")
}

func (e *exchange) ClosePosition(position *models.Position) (*models.Position, error) {
	panic("implement me")
}
