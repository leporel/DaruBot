package bitfinex

import (
	"DaruBot/internal/config"
	exchanges2 "DaruBot/internal/core/exchanges"
	logger2 "DaruBot/internal/logger"
	"DaruBot/internal/models"
	"DaruBot/internal/models/exchanges"
	"DaruBot/internal/models/exchanges/bitfinex"
	"DaruBot/pkg/errors"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/tools"
	"DaruBot/pkg/watcher"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/balanceinfo"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/notification"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/order"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/position"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/ticker"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/trade"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/tradeexecution"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/tradeexecutionupdate"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/wallet"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	supportEventsBitfinex = watcher.EventsMap{
		models.EventError,
		models.EventRequestSuccess,
		models.EventRequestFail,

		models.EventTickerState,
		models.EventCandleState,

		models.EventOrderNew,
		models.EventOrderFilled,
		models.EventOrderCancel,
		models.EventOrderPartiallyFilled,
		models.EventOrderUpdate,

		models.EventPositionNew,
		models.EventPositionClosed,
		models.EventPositionUpdate,

		models.EventWalletUpdate,
	}
)

type Bitfinex struct {
	ctx context.Context

	ws   *websocket.Client
	rest *rest.Client

	log logger.Logger
	cfg config.Configurations

	ready          bool
	readyChan      chan interface{}
	disconnectChan chan interface{}

	orders    *bitfinex.BitfinexOrders
	positions *bitfinex.BitfinexPositions

	subscriptions   models.Subscriptions
	walletsExchange models.Wallets
	walletsMargin   models.Wallets
	balance         models.BalanceUSD

	lastUpdate time.Time

	watchers *watcher.Manager
}

func NewBitfinex(ctx context.Context, c config.Configurations, wManager *watcher.Manager, lg logger.Logger) (exchanges2.Exchange, error) {
	return newBitfinex(ctx, c, wManager, lg)
}

func newBitfinex(ctx context.Context, c config.Configurations, wManager *watcher.Manager, lg logger.Logger) (*Bitfinex, error) {
	p := websocket.NewDefaultParameters()
	p.ManageOrderbook = false
	p.LogTransport = false

	p.ResubscribeOnReconnect = true
	p.AutoReconnect = true
	p.ReconnectAttempts = 1
	p.ReconnectInterval = time.Second * 3

	log := logging.MustGetLogger("Bitfinex_internal")
	logging.SetLevel(logging.INFO, log.Module)
	//if !c.IsDebug() {
	//	logging.SetLevel(logging.ERROR, "Bitfinex_internal")
	//}
	log.SetBackend(logger2.ConvertToGoLogging(lg.WithPrefix("exchange", log.Module), logging.INFO))
	p.Logger = log

	WebSocket := websocket.NewWithParams(p).Credentials(c.Exchanges.Bitfinex.ApiKey, c.Exchanges.Bitfinex.ApiSec)
	REST := rest.NewClient().Credentials(c.Exchanges.Bitfinex.ApiKey, c.Exchanges.Bitfinex.ApiSec)

	status, err := REST.Platform.Status()
	if err != nil || !status {
		return nil, exchanges2.ErrNotOperate
	}

	err = wManager.RegisterEvents(exchanges.ExchangeTypeBitfinex.String(), supportEventsBitfinex)
	if err != nil {
		return nil, err
	}

	return &Bitfinex{
		ctx:             ctx,
		ws:              WebSocket,
		rest:            REST,
		log:             lg.WithPrefix("exchange", "Bitfinex"),
		walletsExchange: models.Wallets{WalletType: models.WalletTypeExchange},
		walletsMargin:   models.Wallets{WalletType: models.WalletTypeMargin},
		balance:         models.BalanceUSD{},
		subscriptions:   models.Subscriptions{},
		orders:          &bitfinex.BitfinexOrders{},
		positions:       &bitfinex.BitfinexPositions{},
		readyChan:       make(chan interface{}, 1),
		disconnectChan:  make(chan interface{}, 1),
		watchers:        wManager,
		cfg:             c,
	}, nil
}

func (b *Bitfinex) Connect() error {
	if b.ready || b.ws.IsConnected() {
		return nil
	}

	err := b.ws.Connect()
	if err != nil {
		b.log.Error("could not connect", err)
		return err
	}

	b.readyChan = make(chan interface{}, 1)

	errorPipe, err := b.watchers.New("bf_api_errors", models.EventError)
	if err != nil {
		return err
	}
	defer b.watchers.Remove("bf_api_errors")

	go b.listen()

	for {
		select {
		case evt := <-errorPipe.Listen():
			return evt.Payload.(error)
		case <-b.readyChan:
			return nil
		}
	}
}

func (b *Bitfinex) Disconnect() {
	b.disconnectChan <- struct{}{}
}

func (b *Bitfinex) IsReady() bool {
	return b.ready
}

func (b *Bitfinex) Ready() <-chan interface{} {
	return b.readyChan
}

func (b *Bitfinex) SupportEvents() watcher.EventsMap {
	return b.watchers.SupportEvents(string(exchanges.ExchangeTypeBitfinex))
}

func (b *Bitfinex) listen() {
	defer func() {
		b.ready = false
		b.ws.Close()
		b.log.Info("websocket disconnected")
	}()

	defer tools.Recover(b.log)

	events := b.ws.Listen()

	for {
		select {
		case obj := <-events:
			switch data := obj.(type) {
			case *websocket.AuthEvent:
				b.log.Info("websocket authorization complete")

				b.ready = true
				close(b.readyChan)

				// TODO check permissions data.Caps needed for bot operations

			case *websocket.InfoEvent:
				// this event confirms connection to the bfx websocket
				b.log.Debugf("INFO EVENT: %#v", data)

				if data.Platform.Status == 0 {
					b.log.Error(exchanges2.ErrNotOperate)
					b.emmit(models.EventError, exchanges2.ErrNotOperate)
					return
				}

			case websocket.PlatformInfo:
				b.log.Debugf("PLATFORM INFO: %#v", data)
				if data.Status == 0 {
					b.log.Error(exchanges2.ErrNotOperate)
					b.emmit(models.EventError, exchanges2.ErrNotOperate)
					return
				}

			case *websocket.SubscribeEvent:
				b.log.Debugf("SUBSCRIBE EVENT %#v", data)

			case *websocket.UnsubscribeEvent:
				b.log.Debugf("UNSUBSCRIBE EVENT %#v", data)

			case *wallet.Snapshot:
				b.log.Debugf("WALLET SNAPSHOT %#v", data)

				b.walletsMargin.Clear()
				b.walletsExchange.Clear()

				for _, w := range data.Snapshot {
					b.updateWallet(w)
				}
				b.lastUpdate = time.Now()

			case *wallet.Update:
				b.log.Debugf("WALLET UPDATE %#v", data)

				w := wallet.Wallet(*data)
				wl := b.updateWallet(&w)
				b.emmit(models.EventWalletUpdate, *wl)

				b.lastUpdate = time.Now()

			case *balanceinfo.Update:
				b.log.Debugf("BALANCE INFO %#v", data)

				b.balance.Total = data.TotalAUM
				b.balance.Total = data.NetAUM

				b.lastUpdate = time.Now()

			case *position.Snapshot:
				b.log.Debugf("POSITION SNAPSHOT %#v", data)

				for _, p := range data.Snapshot {
					b.processPosition(p)
				}
				b.lastUpdate = time.Now()

			case *position.Update:
				b.log.Debugf("POSITION UPDATE %#v", data)

				b.processPosition((*position.Position)(data))
				b.lastUpdate = time.Now()

			case *position.New:
				b.log.Debugf("POSITION NEW %#v", data)
				p := (*position.Position)(data)

				b.positions.Add(p)
				b.lastUpdate = time.Now()

				b.emmit(models.EventPositionNew, *b.convertPosition(data))

			case *position.Cancel:
				b.log.Debugf("POSITION CANCEL %#v", data)

				b.processPosition((*position.Position)(data))
				b.lastUpdate = time.Now()

			case *order.Snapshot:
				b.log.Debugf("ORDER SNAPSHOT %#v", data)

				for _, o := range data.Snapshot {
					b.processOrder(o)
				}

				b.lastUpdate = time.Now()

			case *order.Update:
				b.log.Debugf("ORDER UPDATE %#v", data)

				b.processOrder((*order.Order)(data))

				b.lastUpdate = time.Now()

			case *order.Cancel:
				b.log.Debugf("ORDER CANCEL %#v", data)

				b.processOrder((*order.Order)(data))

				b.lastUpdate = time.Now()

			case *order.New:
				b.log.Debugf("ORDER NEW %#v", data)

				o := (*order.Order)(data)
				b.orders.Add(o)
				b.emmit(models.EventOrderNew, *b.convertOrder(o))

				b.lastUpdate = time.Now()

			case *tradeexecution.TradeExecution:
				b.log.Debugf("TRADE EXECUTION:  %#v", data)

			case *tradeexecutionupdate.TradeExecutionUpdate:
				b.log.Debugf("TRADE EXECUTION UPDATE:  %#v", data)

			case *trade.Trade:
				b.log.Debugf("TRADE NEW:  %#v", data)

			case *trade.Snapshot:
				b.log.Debugf("TRADE SNAPSHOT:  %#v", data)

			case *ticker.Snapshot:
				b.log.Debugf("TICKER SNAPSHOT:  %#v", data)

				for _, t := range data.Snapshot {
					b.emmit(models.EventTickerState, *b.convertTicker(t))
				}

			case *ticker.Ticker:
				b.log.Debugf("TICKER:  %#v", data)

				b.emmit(models.EventTickerState, *b.convertTicker(data))

			case *ticker.Update:
				b.log.Debugf("TICKER UPDATE:  %#v", data)

				b.emmit(models.EventTickerState, *b.convertTicker(data))

			case *candle.Snapshot:
				b.log.Debugf("CANDLE SNAPSHOT:  %#v", data)

				for _, c := range data.Snapshot {
					b.emmit(models.EventTickerState, *b.convertCandle(c))
				}

			case *candle.Candle:
				b.log.Debugf("CANDLE:  %#v", data)

				b.emmit(models.EventTickerState, *b.convertCandle(data))

			case *notification.Notification:
				b.log.Debugf("NOTIFICATION NEW:  %#v", data)

				ord := &order.Order{}

				switch t := data.NotifyInfo.(type) {
				case *order.Order:
					ord = t
				case *order.New:
					ord = (*order.Order)(t)
				case *order.Cancel:
					ord = (*order.Order)(t)
				}

				ordID := ""

				switch data.Type {
				case "oc-req":
					if ord != nil {
						ordID = fmt.Sprint(ord.ID)
					}
				case "on-req":
					if ord != nil {
						ordID = fmt.Sprint(ord.CID)
					}
				}

				meta := make(map[string]string)
				if ordID != "" {
					meta["order_id"] = ordID
				}

				switch data.Status {
				case "ERROR", "FAILURE":
					b.log.Warnf("REQUEST ERROR:  %#v", data)
					b.emmit(models.EventRequestFail, models.RequestResult{Msg: data.Text, Err: exchanges2.ErrRequestError, Meta: meta})
				case "SUCCESS":
					b.emmit(models.EventRequestSuccess, models.RequestResult{Msg: data.Text, Meta: meta})
				default:
					b.log.Warnf("UNKNOWN NOTIFICATION:  %#v", data)
				}

			case error:
				err := errors.WrapMessage(exchanges2.ErrWebsocketError, fmt.Sprintf("channel closed: %s", data.Error()))
				b.log.Error(err)
				b.emmit(models.EventError, err)
				return

			default:
				b.log.Debugf("MSG RECV: %#v", data)
			}

		case <-b.disconnectChan:
			b.log.Debugf("disconnect from web socket")
			return

		case <-b.ctx.Done():
			b.log.Debugf("gracefully stop received")
			return
		}
	}
}

func (b *Bitfinex) processOrder(o *order.Order) {
	if strings.Contains(o.Status, "EXECUTED") {
		b.orders.Delete(o.ID)
		b.emmit(models.EventOrderFilled, *b.convertOrder(o))
	}
	if strings.Contains(o.Status, "CANCELED") {
		b.orders.Delete(o.ID)
		b.emmit(models.EventOrderCancel, *b.convertOrder(o))
	}
	if strings.Contains(o.Status, "PARTIALLY FILLED") {
		b.orders.Add(o)
		b.emmit(models.EventOrderPartiallyFilled, *b.convertOrder(o))
	}
	if strings.Contains(o.Status, "ACTIVE") {
		b.orders.Add(o)
		b.emmit(models.EventOrderUpdate, *b.convertOrder(o))
	}
}

func (b *Bitfinex) processPosition(p *position.Position) {
	if p.Status != "CLOSED" {
		b.positions.Add(p)
		b.emmit(models.EventPositionUpdate, *b.convertPosition(p))
	}
	if p.Status == "CLOSED" {
		b.positions.Delete(p.Id)
		b.emmit(models.EventPositionClosed, *b.convertPosition(p))
	}
}

func (b *Bitfinex) emmit(eventHead watcher.EventHead, data interface{}) {
	err := b.watchers.Emmit(watcher.BuildEvent(eventHead, string(exchanges.ExchangeTypeBitfinex), data))
	if err != nil {
		b.log.Error(err)
	}
}

/*
	Subscribes
*/

func (b *Bitfinex) SubscribeTicker(pair string) (string, error) {
	sid, err := b.ws.SubscribeTicker(b.ctx, pair)
	if err != nil {
		return "", err
	}
	b.subscriptions.Add(&models.Subscription{
		ID:   sid,
		Pair: pair,
		Type: models.SubTypeTicker,
	})
	return sid, nil
}

func (b *Bitfinex) SubscribeCandles(pair string, resolution models.CandleResolution) (string, error) {
	cres, err := candleResolutionToBitfinex(resolution)
	if err != nil {
		return "", err
	}

	sid, err := b.ws.SubscribeCandles(b.ctx, pair, cres)
	if err != nil {
		return "", err
	}

	b.subscriptions.Add(&models.Subscription{
		ID:   sid,
		Pair: pair,
		Type: models.SubTypeCandle,
	})

	return sid, nil
}

//func (b *Bitfinex) SubscribeStatus(pair string) (string, error) {
//	sid, err := b.ws.SubscribeStatus(b.ctx, "global", "liq")
//	if err != nil {
//		return "", err
//	}
//
//	b.subscriptions.Add(&models.Subscription{
//		ID:   sid,
//		Pair: pair,
//		Type: models.SubTypeCandle,
//	})
//
//	return sid, nil
//}

func (b *Bitfinex) Unsubscribe(sid string) error {
	err := b.ws.Unsubscribe(b.ctx, sid)
	b.subscriptions.Delete(sid)
	return err
}

func (b *Bitfinex) GetSubscriptions() *models.Subscriptions {
	return &b.subscriptions
}

/*
	Data
*/

func (b *Bitfinex) HasUpdates(t time.Time) bool {
	return t.Before(b.lastUpdate)
}

func (b *Bitfinex) GetOrders() ([]*models.Order, error) {
	rs := make([]*models.Order, 0)

	orders := b.orders.GetAll()

	for _, o := range orders {
		rs = append(rs, b.convertOrder(o))
	}

	return rs, nil
}

func (b *Bitfinex) GetPositions() ([]*models.Position, error) {
	rs := make([]*models.Position, 0)

	ps := b.positions.GetAll()

	for _, p := range ps {
		rs = append(rs, b.convertPosition(p))
	}

	return rs, nil
}

func (b *Bitfinex) GetWallets() ([]*models.Wallets, error) {
	return []*models.Wallets{&b.walletsExchange, &b.walletsMargin}, nil
}

func (b *Bitfinex) updateWallet(w *wallet.Wallet) *models.WalletCurrency {
	wl := &models.WalletCurrency{
		Name:      w.Currency,
		Available: w.BalanceAvailable,
		Balance:   w.Balance,
	}

	switch w.Type {
	case "exchange":
		wl.WalletType = models.WalletTypeExchange
		b.walletsExchange.Add(wl)
	case "margin":
		wl.WalletType = models.WalletTypeMargin
		b.walletsMargin.Add(wl)
	}

	return wl
}

func (b *Bitfinex) GetBalance() (models.BalanceUSD, error) {
	return b.balance, nil
}

/*
	Requests
*/

func (b *Bitfinex) GetTicker(pair string) (*models.Ticker, error) {
	t, err := b.rest.Tickers.Get(pair)
	if err != nil {
		return nil, err
	}
	return b.convertTicker(t), nil
}

// https://docs.bitfinex.com/reference#rest-public-candles
func (b *Bitfinex) GetCandles(pair string, resolution models.CandleResolution, start time.Time, end time.Time) (*models.Candles, error) {
	cres, err := candleResolutionToBitfinex(resolution)
	if err != nil {
		return nil, err
	}

	cs, err := b.rest.Candles.HistoryWithQuery(
		pair,
		cres,
		common.Mts(tools.TimeToMilliseconds(start)),
		common.Mts(tools.TimeToMilliseconds(end)),
		1000, // Max 10000
		1,
	)
	if err != nil {
		return nil, err
	}

	rs := &models.Candles{
		Pair:       pair,
		Resolution: resolution,
		Candles:    make([]*models.Candle, 0, len(cs.Snapshot)),
	}

	for _, c := range cs.Snapshot {
		rs.Candles = append(rs.Candles, b.convertCandle(c))
	}

	return rs, nil
}

// PutOrder https://docs.bitfinex.com/reference#ws-auth-input-order-new
func (b *Bitfinex) PutOrder(o *models.PutOrder) (*models.Order, error) {
	if !b.ready {
		return nil, exchanges2.ErrNoConnect
	}

	var err error

	var typeOrder string
	var PriceAuxLimit float64
	var Price = o.Price
	var Amount = o.Amount
	var Pair = o.Pair
	var orderClientID int64

	switch o.Type {
	case models.OrderTypeLimit:
		typeOrder = "LIMIT"
	case models.OrderTypeMarket:
		typeOrder = "MARKET"
		Price = 0
	case models.OrderTypeStop:
		typeOrder = "STOP"
		if o.StopPrice == 0 {
			return nil, errors.WrapMessage(exchanges2.ErrInvalidRequestParams, "stop price are not specified")
		}
		Price = o.StopPrice
	case models.OrderTypeStopLimit:
		typeOrder = "STOP LIMIT"
		if o.Price == 0 {
			return nil, errors.WrapMessage(exchanges2.ErrInvalidRequestParams, "limit price are not specified")
		}
		PriceAuxLimit = o.Price
		if o.StopPrice == 0 {
			return nil, errors.WrapMessage(exchanges2.ErrInvalidRequestParams, "stop price are not specified")
		}
		Price = o.StopPrice
	default:
		return nil, exchanges2.ErrOrderTypeNotSupported
	}

	if !o.Margin {
		typeOrder = fmt.Sprintf("EXCHANGE %s", typeOrder)
	}

	if !strings.HasPrefix(Pair, "t") {
		return nil, exchanges2.ErrPairIncorrect
	}

	if o.InternalID != "" {
		orderClientID, err = strconv.ParseInt(o.InternalID, 10, 64)
		if err != nil {
			return nil, errors.WrapMessage(exchanges2.ErrInvalidRequestParams, err)
		}
	} else {
		o.InternalID = fmt.Sprint(time.Now().Unix() / 1000)
	}

	req := &order.NewRequest{
		GID:           0,
		CID:           orderClientID,
		Type:          typeOrder,
		Symbol:        Pair,
		Amount:        Amount,
		Price:         Price,
		PriceAuxLimit: PriceAuxLimit,
		AffiliateCode: b.cfg.Exchanges.Bitfinex.Affiliate(),
	}

	b.log.Debugf("Submitting order: %#v", req)
	err = b.ws.SubmitOrder(b.ctx, req)
	if err != nil {
		return nil, err
	}

	orderPipe, err := b.watchers.New(fmt.Sprint("bf_wait_order", orderClientID), models.EventOrderFilled, models.EventOrderNew, models.EventRequestFail)
	if err != nil {
		return nil, err
	}
	defer b.watchers.Remove(fmt.Sprint("bf_wait_order", orderClientID))

	Timout := time.NewTimer(3 * time.Second)
	defer Timout.Stop()

	for {
		select {
		case evt := <-orderPipe.Listen():
			switch {
			case evt.Is(models.EventRequestFail):
				rr := evt.Payload.(models.RequestResult)
				if rr.Meta["order_id"] == o.InternalID {
					return nil, errors.WrapMessage(rr.Err, rr.Msg)
				}
			case evt.Is(models.EventOrderFilled), evt.Is(models.EventOrderNew):
				or := evt.Payload.(models.Order)
				if or.InternalID == o.InternalID {
					return &or, nil
				}
			}

		case <-Timout.C:
			return nil, exchanges2.ErrResultTimeOut
		}
	}
}

func (b *Bitfinex) CancelOrder(o *models.Order) error {
	if !b.ready {
		return exchanges2.ErrNoConnect
	}

	id, err := strconv.ParseInt(o.ID, 10, 64)
	if err != nil {
		return errors.WrapMessage(exchanges2.ErrInvalidRequestParams, err)
	}

	var cid int64 = 0
	date := "" // 2016-12-05
	if id == 0 {
		cid = o.GetInternalIDAsInt()
		date = o.Date.Format("2006-01-02")
	}

	req := order.CancelRequest{
		ID:      id,
		CID:     cid,
		CIDDate: date,
	}

	b.log.Debugf("Canceling order: %#v", req)
	err = b.ws.SubmitCancel(b.ctx, &req)
	if err != nil {
		return err
	}

	orderPipe, err := b.watchers.New(fmt.Sprint("bf_wait_order", req.ID, req.CID), models.EventOrderCancel, models.EventRequestFail)
	if err != nil {
		return err
	}
	defer b.watchers.Remove(fmt.Sprint("bf_wait_order", req.ID, req.CID))

	Timout := time.NewTimer(3 * time.Second)
	defer Timout.Stop()

	for {
		select {
		case evt := <-orderPipe.Listen():
			switch {
			case evt.Is(models.EventRequestFail):
				rr := evt.Payload.(models.RequestResult)
				if rr.Meta["order_id"] == o.ID || rr.Meta["order_id"] == o.InternalID {
					return errors.WrapMessage(rr.Err, rr.Msg)
				}
			case evt.Is(models.EventOrderCancel):
				or := evt.Payload.(models.Order)
				if or.ID == o.ID || or.InternalID == o.InternalID {
					return nil
				}
			}

		case <-Timout.C:
			return exchanges2.ErrResultTimeOut
		}
	}
}

// UpdateOrder if price, priceStop, amount equals 0, they will be ignored
func (b *Bitfinex) UpdateOrder(orderID string, price float64, priceStop float64, amount float64) (*models.Order, error) {
	if !b.ready {
		return nil, exchanges2.ErrNoConnect
	}

	id, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, errors.WrapMessage(exchanges2.ErrInvalidRequestParams, err)
	}

	o := b.orders.Get(id)
	if o == nil {
		return nil, exchanges2.ErrOrderNotFound
	}

	req := &order.UpdateRequest{
		ID:     id,
		Price:  price,
		Amount: amount,
	}

	if priceStop != 0 {
		if !strings.Contains(o.Type, "STOP LIMIT") {
			return nil, errors.WrapMessage(exchanges2.ErrInvalidRequestParams, "order is not STOP LIMIT type")
		}

		req.PriceAuxLimit = price
		req.Price = priceStop
	}

	b.log.Debugf("Updating order: %#v", req)
	err = b.ws.SubmitUpdateOrder(b.ctx, req)
	if err != nil {
		return nil, err
	}

	orderPipe, err := b.watchers.New(fmt.Sprint("bf_wait_order", id), models.EventOrderUpdate)
	if err != nil {
		return nil, err
	}
	defer b.watchers.Remove(fmt.Sprint("bf_wait_order", id))

	Timout := time.NewTimer(3 * time.Second)
	defer Timout.Stop()

	for {
		select {
		case evt := <-orderPipe.Listen():
			switch {
			case evt.Is(models.EventOrderUpdate):
				ord := evt.Payload.(models.Order)
				if ord.ID == orderID {
					return &ord, nil
				}
			}

		case <-Timout.C:
			return nil, exchanges2.ErrResultTimeOut
		}
	}
}

func (b *Bitfinex) ClosePosition(p *models.Position) (*models.Position, error) {
	if !b.ready {
		return nil, exchanges2.ErrNoConnect
	}

	var Pair = p.Pair

	var prevStatePos = models.Position{}
	pos := b.positions.Get(p.GetIDAsInt())
	if pos == nil {
		return nil, exchanges2.ErrPositionNotFound
	}
	prevStatePos = *b.convertPosition(pos)

	req := &order.NewRequest{
		CID:           time.Now().Unix() / 1000,
		Symbol:        Pair,
		Type:          "MARKET",
		Amount:        -p.Amount,
		AffiliateCode: b.cfg.Exchanges.Bitfinex.Affiliate(),
		Close:         true,
	}

	b.log.Debugf("Submitting order to close position: %#v", req)
	err := b.ws.SubmitOrder(b.ctx, req)
	if err != nil {
		return nil, err
	}

	positionPipe, err := b.watchers.New(fmt.Sprint("bf_wait_order", req.CID), models.EventPositionClosed, models.EventRequestFail)
	if err != nil {
		return nil, err
	}
	defer b.watchers.Remove(fmt.Sprint("bf_wait_order", req.CID))

	Timout := time.NewTimer(3 * time.Second)
	defer Timout.Stop()

	for {
		select {
		case evt := <-positionPipe.Listen():
			switch {
			case evt.Is(models.EventRequestFail):
				rr := evt.Payload.(models.RequestResult)
				if rr.Meta["order_id"] == fmt.Sprint(req.CID) {
					return nil, errors.WrapMessage(rr.Err, rr.Msg)
				}
			case evt.Is(models.EventPositionClosed):
				pos := evt.Payload.(models.Position)
				if pos.ID == p.ID {
					return &prevStatePos, nil
				}
			}

		case <-Timout.C:
			return nil, exchanges2.ErrResultTimeOut
		}
	}
}

func (b *Bitfinex) CheckPair(pair string, margin bool) error {
	return checkPair(pair, margin)
}

func checkPair(pair string, margin bool) error {
	var pairs [][]string

	if !strings.HasPrefix(pair, "t") {
		return exchanges2.ErrPairIncorrect
	}

	pair = strings.TrimPrefix(pair, "t")

	url := "https://api-pub.bitfinex.com/v2/conf/pub:list:pair:exchange"

	if margin {
		url = "https://api-pub.bitfinex.com/v2/conf/pub:list:pair:margin"
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &pairs)
	if err != nil {
		return err
	}

	for _, p := range pairs[0] {
		if p == pair {
			return nil
		}
	}

	return exchanges2.ErrPairNotSupported
}
