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
		models.EventError, // TODO

		models.EventTickerState, // TODO
		models.EventCandleState, // TODO

		models.EventOrderNew,             // TODO
		models.EventOrderFilled,          // TODO
		models.EventOrderCancel,          // TODO
		models.EventOrderPartiallyFilled, // TODO
		models.EventOrderUpdate,          // TODO

		models.EventPositionNew,    // TODO
		models.EventPositionClosed, // TODO
		models.EventPositionUpdate, // TODO

		models.EventWalletUpdate, // TODO
	}
)

type (
	quoteFunc = func(symbol string, startDate, endDate string, period quote.Period) (quote.Quote, error)
)

func getCandle(q *quote.Quote, t time.Time) *models.Candle {
	if len(q.Date) == 0 {
		return nil
	}

	if t.IsZero() {
		var d time.Duration
		if len(q.Date) > 1 {
			d = q.Date[1].Sub(q.Date[0])
		}
		return models.QuoteToModel(q, q.Symbol, len(q.Date)-1, d)
	}

	if len(q.Date) < 2 {
		return nil
	}

	d := q.Date[1].Sub(q.Date[0])

	for i := 0; i < len(q.Date); i++ {
		if q.Date[i].Sub(t) <= d {
			return models.QuoteToModel(q, q.Symbol, i, d)
		}
	}

	return nil
}

type exchange struct {
	market    string
	quoteFunc quoteFunc

	lastUpdate time.Time

	ctx context.Context
	log logger.Logger
	cfg config.Configurations

	ready     bool
	readyChan chan interface{}

	watchers *watcher.Manager
	cache    *cache.Cache

	dio *theWorld
}

func NewExchangeMock(ctx context.Context,
	wManager *watcher.Manager,
	lg logger.Logger,
	cfg config.Configurations,
	market string, quoteF quoteFunc,
	from, to time.Time) (exchanges2.CryptoExchange, error) {
	return newExchangeMock(ctx, wManager, lg, cfg, market, quoteF, from, to)
}

func newExchangeMock(ctx context.Context,
	wManager *watcher.Manager,
	lg logger.Logger,
	cfg config.Configurations,
	market string, quoteF quoteFunc,
	from, to time.Time) (*exchange, error) {

	if !quote.ValidMarket(market) {
		return nil, errors.New("market not supported")
	}

	err := wManager.RegisterEvents(exchanges.ExchangeTypeMock.String(), supportEvents)
	if err != nil {
		return nil, err
	}

	c := cache.New(10*time.Minute, 15*time.Minute)

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
		dio:        newTheWorld(from, to),
	}

	return rs, nil
}

func (e *exchange) getQuote(pair string, period quote.Period) (*quote.Quote, error) {
	curTime := e.dio.CurrentTime()

	key := ""
	var from, to time.Time

	switch period {
	case quote.Daily:
		from = time.Date(e.dio.from.Year(), e.dio.from.Month(), e.dio.from.Day()-1, 0, 0, 0, 0, time.Local)
		to = time.Date(e.dio.to.Year(), e.dio.to.Month(), e.dio.to.Day(), 23, 59, 59, 0, time.Local)
		key = getDailyKey(e.dio.from, e.dio.to, pair)
	case quote.Min1:
		key = getMinuteKey(curTime, pair)
		from = time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, time.Local)
		to = time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 23, 59, 59, 0, time.Local)
	default:
		return nil, errors.New("period not set")
	}

	e.log.Tracef("cached quote %s", key)

	q, found := e.cache.Get(key)
	if !found {
		qNew, err := e.downloadQuote(from, to, pair, period)
		if err != nil {
			return nil, err
		}

		if len(qNew.Low) == 0 {
			return nil, errors.New("cant get quote")
		}

		q = qNew
		e.cache.Set(key, qNew, cache.DefaultExpiration)
	}

	return q.(*quote.Quote), nil
}

func (e *exchange) downloadQuote(from, to time.Time, pair string, period quote.Period) (*quote.Quote, error) {
	start := from.UTC()
	end := to.UTC()
	//q, err := quote.NewQuoteFromCoinbase(pair, start.Format(timeFormat), end.Format(timeFormat), period)
	q, err := e.quoteFunc(pair, timeString(start), timeString(end), period)
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

func (e *exchange) GetTicker(pair string) (*models.Ticker, error) {
	e.dio.TimeStop()
	defer e.dio.TimeStart()
	curTime := e.dio.CurrentTime()

	e.log.Tracef("get ticker, time: %s (UTC %s)", timeString(curTime), timeString(curTime.UTC()))

	qd, err := e.getQuote(pair, quote.Daily)
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

	qm, err := e.getQuote(pair, quote.Min1)
	if err != nil {
		return nil, err
	}
	candle = getCandle(qm, curTime)
	//e.log.Tracef("minute candle: %+v", candle)

	ticker := &models.Ticker{
		Pair:     pair,
		Price:    getRandFloat(candle.Low, candle.High),
		Exchange: exchanges.ExchangeTypeMock,
		State:    dayState,
	}

	e.log.Tracef("formed ticker: %+v", ticker)

	return ticker, nil
}

func (e *exchange) GetCandles(pair string, resolution models.CandleResolution, start time.Time, end time.Time) (*models.Candles, error) {
	panic("implement me")
}

func (e *exchange) GetSubscriptions() *models.Subscriptions {
	panic("implement me")
}

func (e *exchange) SubscribeTicker(pair string) (subID string, err error) {
	panic("implement me")
}

func (e *exchange) SubscribeCandles(pair string, resolution models.CandleResolution) (subID string, err error) {
	panic("implement me")
}

func (e *exchange) Unsubscribe(subID string) error {
	panic("implement me")
}

func (e *exchange) CheckSymbol(pair string, margin bool) error {
	list, err := quote.NewMarketList(e.market)
	if err != nil {
		return err
	}
	for _, i2 := range list {
		if i2 == pair {
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
