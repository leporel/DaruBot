package mock

import (
	"DaruBot/internal/config"
	exchanges2 "DaruBot/internal/core/exchanges"
	"DaruBot/internal/models"
	"DaruBot/internal/models/exchanges"
	"DaruBot/pkg/errors"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/watcher"
	"context"
	"github.com/markcheno/go-quote"
	"github.com/patrickmn/go-cache"
	"time"
)

var (
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

	watchers *watcher.Manager
	cache    *cache.Cache

	lastUpdate    time.Time
	subscriptions models.Subscriptions
}

func NewExchangeMock(ctx context.Context,
	wManager *watcher.Manager,
	lg logger.Logger,
	cfg config.Configurations,
	market string, quoteF quoteFunc,
	stand *TheWorld) (exchanges2.CryptoExchange, error) {
	return newExchangeMock(ctx, wManager, lg, cfg, market, quoteF, stand)
}

func newExchangeMock(ctx context.Context,
	wManager *watcher.Manager,
	lg logger.Logger,
	cfg config.Configurations,
	market string, quoteF quoteFunc,
	stand *TheWorld) (*exchange, error) {

	if !quote.ValidMarket(market) {
		return nil, errors.New("market not supported")
	}

	err := wManager.RegisterEvents(exchanges.ExchangeTypeMock.String(), supportEvents)
	if err != nil {
		return nil, err
	}

	c := cache.New(10*time.Minute, 0)

	rs := &exchange{
		market:     market,
		quoteFunc:  quoteF,
		lastUpdate: time.Time{},
		ctx:        ctx,
		log:        lg.WithPrefix("exchange", "Mock"),
		cfg:        cfg,
		ready:      false,
		readyChan:  make(chan interface{}, 1),
		watchers:   wManager,
		cache:      c,
		dio:        stand,
	}

	return rs, nil
}

func (e *exchange) getQuote(symbol string, period quote.Period) (*quote.Quote, error) {
	curTime := e.dio.CurrentTime()

	key := ""
	var from, to time.Time

	switch period {
	case quote.Daily:
		from = time.Date(e.dio.from.Year(), e.dio.from.Month(), e.dio.from.Day()-1, 0, 0, 0, 0, time.Local)
		to = time.Date(e.dio.to.Year(), e.dio.to.Month(), e.dio.to.Day(), 23, 59, 59, 0, time.Local)
		key = getDailyKey(e.dio.from, e.dio.to, symbol)
	case quote.Min1:
		key = getMinuteKey(curTime, symbol)
		from = time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, time.Local)
		to = time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 23, 59, 59, 0, time.Local)
	default:
		return nil, errors.New("period not set")
	}

	e.log.Tracef("cached quote %s", key)

	q, found := e.cache.Get(key)
	if !found {
		qNew, err := e.downloadQuote(from, to, symbol, period)
		if err != nil {
			return nil, err
		}

		if len(qNew.Low) == 0 {
			return nil, errors.New("cant get quote")
		}

		q = qNew
		e.cache.Set(key, qNew, cache.NoExpiration)
	}

	return q.(*quote.Quote), nil
}

func (e *exchange) downloadQuote(from, to time.Time, symbol string, period quote.Period) (*quote.Quote, error) {
	start := from.UTC()
	end := to.UTC()
	q, err := e.quoteFunc(symbol, quoteFormat(start), quoteFormat(end), period)
	if err != nil {
		return nil, err
	}
	return &q, nil
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

func (e *exchange) GetTicker(symbol string) (*models.Ticker, error) {
	e.dio.TimeStop()
	defer e.dio.TimeStart()
	curTime := e.dio.CurrentTime()

	e.log.Tracef("get ticker, time: %s (UTC %s)", quoteFormat(curTime), quoteFormat(curTime.UTC()))

	qd, err := e.getQuote(symbol, quote.Daily)
	if err != nil {
		return nil, err
	}

	candle := getCandle(qd, curTime)
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

	qm, err := e.getQuote(symbol, quote.Min1)
	if err != nil {
		return nil, err
	}
	candle = getCandle(qm, curTime)
	//e.log.Tracef("minute candle: %+v", candle)

	ticker := &models.Ticker{
		Symbol:   symbol,
		Price:    getRandFloat(candle.Low, candle.High),
		Exchange: exchanges.ExchangeTypeMock,
		State:    dayState,
	}

	e.log.Tracef("formed ticker: %+v", ticker)

	return ticker, nil
}

func (e *exchange) GetCandles(symbol string, resolution models.CandleResolution, from time.Time, to time.Time) (*models.Candles, error) {
	e.dio.TimeStop()
	defer e.dio.TimeStart()

	res, err := resolution.ToQuoteModel()
	if err != nil {
		return nil, err
	}

	e.log.Tracef("get candles, from %s to %s", quoteFormat(from), quoteFormat(to))

	if !from.IsZero() && to.After(from) {
		q, err := e.downloadQuote(from, to, symbol, res)
		if err != nil {
			return nil, err
		}
		return models.QuoteToModels(q, symbol), nil
	}

	return nil, exchanges2.ErrInvalidRequestParams
}

func (e *exchange) GetLastCandle(symbol string, resolution models.CandleResolution) (*models.Candle, error) {
	e.dio.TimeStop()
	defer e.dio.TimeStart()

	res, err := resolution.ToQuoteModel()
	if err != nil {
		return nil, err
	}

	e.log.Tracef("get last candle, from %s", quoteFormat(e.dio.CurrentTime()))

	q, err := e.downloadQuote(e.dio.CurrentTime().Add(-resolution.ToDuration()), e.dio.CurrentTime(), symbol, res)
	if err != nil {
		return nil, err
	}
	//e.log.Tracef("quote: %+v", q)

	return models.QuoteToModel(q, symbol, len(q.Date)-1, resolution), nil
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
