package exchanges

import (
	"DaruBot/internal/config"
	logger2 "DaruBot/internal/logger"
	"DaruBot/internal/models"
	"DaruBot/internal/models/exchanges"
	"DaruBot/pkg/errors"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/tools"
	"DaruBot/pkg/watcher"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/balanceinfo"
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
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	supportEventsBitfinex = watcher.EventsMap{
		EventError:          reflect.TypeOf((*error)(nil)).Elem(),
		EventRequestSuccess: reflect.TypeOf(&RequestResult{}).Elem(),
		EventRequestFail:    reflect.TypeOf(&RequestResult{}).Elem(),

		EventOrderNew:             reflect.TypeOf(&models.Order{}).Elem(),
		EventOrderFilled:          reflect.TypeOf(&models.Order{}).Elem(),
		EventOrderCancel:          reflect.TypeOf(&models.Order{}).Elem(),
		EventOrderPartiallyFilled: reflect.TypeOf(&models.Order{}).Elem(),
		EventOrderUpdate:          reflect.TypeOf(&models.Order{}).Elem(),

		EventPositionNew:    reflect.TypeOf(&models.Position{}).Elem(),
		EventPositionClosed: reflect.TypeOf(&models.Position{}).Elem(),
		EventPositionUpdate: reflect.TypeOf(&models.Position{}).Elem(),

		EventWalletUpdate: reflect.TypeOf(&models.WalletCurrency{}).Elem(),
	}
)

type BitFinex struct {
	ctx context.Context

	ws   *websocket.Client
	rest *rest.Client

	lg  logger.Logger
	cfg config.Configurations

	ready          bool
	readyChan      chan interface{}
	disconnectChan chan interface{}

	orders    *exchanges.BitfinexOrders
	positions *exchanges.BitfinexPositions

	walletsExchange models.Wallets
	walletsMargin   models.Wallets
	balance         models.BalanceUSD

	lastUpdate time.Time

	watchers *watcher.Manager
}

func NewBitFinex(ctx context.Context, c config.Configurations, lg logger.Logger) (*BitFinex, error) {
	p := websocket.NewDefaultParameters()
	p.ManageOrderbook = false
	p.LogTransport = false

	log := logging.MustGetLogger("Bitfinex_internal")
	logging.SetLevel(logging.INFO, log.Module)
	//if !c.IsDebug() {
	//	logging.SetLevel(logging.ERROR, "Bitfinex_internal")
	//}
	log.SetBackend(logger2.ConvertToGoLogging(lg.WithPrefix("exchange", log.Module), logging.INFO))
	p.Logger = log

	WebSocket := websocket.NewWithParams(p).Credentials(c.Exchanges.BitFinex.ApiKey, c.Exchanges.BitFinex.ApiSec)
	REST := rest.NewClient().Credentials(c.Exchanges.BitFinex.ApiKey, c.Exchanges.BitFinex.ApiSec)

	wManager := watcher.NewWatcherManager()
	wManager.RegisterEvents(supportEventsBitfinex)

	return &BitFinex{
		ctx:             ctx,
		ws:              WebSocket,
		rest:            REST,
		lg:              lg.WithPrefix("exchange", "Bitfinex"),
		walletsExchange: models.Wallets{WalletType: models.WalletTypeExchange},
		walletsMargin:   models.Wallets{WalletType: models.WalletTypeMargin},
		balance:         models.BalanceUSD{},
		orders:          &exchanges.BitfinexOrders{},
		positions:       &exchanges.BitfinexPositions{},
		readyChan:       make(chan interface{}, 1),
		disconnectChan:  make(chan interface{}, 1),
		watchers:        wManager,
		cfg:             c,
	}, nil
}

func (b *BitFinex) Connect() error {
	if b.ready || b.ws.IsConnected() {
		return nil
	}

	err := b.ws.Connect()
	if err != nil {
		b.lg.Error("could not connect", err)
		return err
	}

	b.readyChan = make(chan interface{}, 1)

	errorPipe, err := b.watchers.New("bf_api_errors", EventError)
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

func (b *BitFinex) Disconnect() {
	b.disconnectChan <- struct{}{}
}

func (b *BitFinex) IsReady() bool {
	return b.ready
}

func (b *BitFinex) Ready() <-chan interface{} {
	return b.readyChan
}

func (b *BitFinex) EventsList() watcher.EventsMap {
	return b.watchers.SupportEvents()
}

func (b *BitFinex) listen() {
	defer func() {
		b.ready = false
		b.ws.Close()
		b.lg.Info("websocket disconnected")
	}()

	defer tools.Recover(b.lg)

	events := b.ws.Listen()

	for {
		select {
		case obj := <-events:
			switch data := obj.(type) {
			case *websocket.AuthEvent:
				b.lg.Info("websocket authorization complete")

				b.ready = true
				close(b.readyChan)

				// TODO check permissions data.Caps needed for bot operations

			case *websocket.InfoEvent:
				// this event confirms connection to the bfx websocket
				b.lg.Debugf("INFO EVENT: %#v", data)

				// TODO check PlatformInfo.Status https://docs.bitfinex.com/reference#rest-public-platform-status

			case *websocket.SubscribeEvent:
				// TODO handle

			case *websocket.UnsubscribeEvent:
				// TODO handle

			case *wallet.Snapshot:
				b.lg.Debugf("WALLET SNAPSHOT %#v", data)

				b.walletsMargin.Clear()
				b.walletsExchange.Clear()

				for _, w := range data.Snapshot {
					b.updateWallet(w)
				}
				b.lastUpdate = time.Now()

			case *wallet.Update:
				b.lg.Debugf("WALLET UPDATE %#v", data)

				w := wallet.Wallet(*data)
				wl := b.updateWallet(&w)
				b.emmit(EventWalletUpdate, *wl)

				b.lastUpdate = time.Now()

			case *balanceinfo.Update:
				b.lg.Debugf("BALANCE INFO %#v", data)

				b.balance.Total = data.TotalAUM
				b.balance.Total = data.NetAUM

				b.lastUpdate = time.Now()

			case *position.Snapshot:
				b.lg.Debugf("POSITION SNAPSHOT %#v", data)

				b.positions.Clear()

				for _, p := range data.Snapshot {
					b.positions.Add(p)
				}
				b.lastUpdate = time.Now()

			case *position.Update:
				b.lg.Debugf("POSITION UPDATE %#v", data)

				if data.Status != "CLOSED" {
					p := position.Position(*data)
					b.positions.Add(&p)
				}
				b.lastUpdate = time.Now()

				b.emmit(EventPositionUpdate, *b.convertPosition(data))

			case *position.New:
				b.lg.Debugf("POSITION NEW %#v", data)
				p := position.Position(*data)

				b.positions.Add(&p)
				b.lastUpdate = time.Now()

				b.emmit(EventPositionNew, *b.convertPosition(data))

			case *position.Cancel:
				b.lg.Debugf("POSITION CANCEL %#v", data)

				b.positions.Delete(data.Id)
				b.lastUpdate = time.Now()

				b.emmit(EventPositionClosed, *b.convertPosition(data))

			case *order.Snapshot:
				b.lg.Debugf("ORDER SNAPSHOT %#v", data)

				b.orders.Clear()

				for _, p := range data.Snapshot {
					b.orders.Add(p)
				}
				b.lastUpdate = time.Now()

			case *order.Update:
				b.lg.Debugf("ORDER UPDATE %#v", data)

				o := order.Order(*data)
				b.orders.Add(&o)

				b.lastUpdate = time.Now()

				if strings.Contains(data.Status, "PARTIALLY FILLED") {
					b.emmit(EventOrderPartiallyFilled, *b.convertOrder(data))
				}

				if strings.Contains(data.Status, "ACTIVE") {
					b.emmit(EventOrderUpdate, *b.convertOrder(data))
				}

			case *order.Cancel:
				b.lg.Debugf("ORDER CANCEL %#v", data)

				b.orders.Delete(data.ID)
				b.lastUpdate = time.Now()

				if strings.Contains(data.Status, "EXECUTED") {
					b.emmit(EventOrderFilled, *b.convertOrder(data))
				}
				if strings.Contains(data.Status, "CANCELED") {
					b.emmit(EventOrderCancel, *b.convertOrder(data))
				}

			case *order.New:
				b.lg.Debugf("ORDER NEW %#v", data)

				o := order.Order(*data)
				b.orders.Add(&o)
				b.lastUpdate = time.Now()

				if data.Status == "ACTIVE" {
					b.emmit(EventOrderNew, *b.convertOrder(data))
				}

			case *tradeexecution.TradeExecution:
				b.lg.Debugf("TRADE EXECUTION:  %#v", data)

			case *tradeexecutionupdate.TradeExecutionUpdate:
				b.lg.Debugf("TRADE EXECUTION UPDATE:  %#v", data)

			case *trade.Trade:
				b.lg.Debugf("TRADE NEW:  %#v", data)

			case *ticker.Snapshot:
				b.lg.Debugf("TICKER NEW:  %#v", data)

			case *ticker.Update:
				b.lg.Debugf("TICKER UPDATE:  %#v", data)

			case *notification.Notification:
				b.lg.Debugf("NOTIFICATION NEW:  %#v", data)

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
					b.lg.Warnf("REQUEST ERROR:  %#v", data)
					b.emmit(EventRequestFail, RequestResult{Msg: data.Text, Err: ErrRequestError, Meta: meta})
				case "SUCCESS":
					b.emmit(EventRequestSuccess, RequestResult{Msg: data.Text, Meta: meta})
				default:
					b.lg.Warnf("UNKNOWN NOTIFICATION:  %#v", data)
				}

			case error:
				b.lg.Errorf("CHANNEL CLOSED:  %#v", data)
				b.emmit(EventError, errors.WrapMessage(ErrWebsocketError, fmt.Sprintf("channel closed: %s", data.Error())))
				return

			default:
				b.lg.Debugf("MSG RECV: %#v", data)
			}

		case <-b.disconnectChan:
			b.lg.Debugf("disconnect from web socket")
			return

		case <-b.ctx.Done():
			b.lg.Debugf("gracefully stop received")
			return
		}
	}
}

func (b *BitFinex) RegisterWatcher(name string, eType ...watcher.EventType) *watcher.Watcher {
	wh, err := b.watchers.New(name, eType...)
	if err != nil {
		b.lg.Error(err)
	}
	return wh
}

func (b *BitFinex) RemoveWatcher(name string) {
	b.watchers.Remove(name)
}

func (b *BitFinex) emmit(EventT watcher.EventType, data interface{}) {
	err := b.watchers.Emmit(watcher.NewEvent(EventT, data))
	if err != nil {
		b.lg.Error(err)
	}
}

func (b *BitFinex) HasUpdates(t time.Time) bool {
	return t.Before(b.lastUpdate)
}

func (b *BitFinex) GetOrders() ([]*models.Order, error) {
	rs := make([]*models.Order, 0)

	orders := b.orders.GetAll()

	for _, o := range orders {
		rs = append(rs, b.convertOrder(o))
	}

	return rs, nil
}

func (b *BitFinex) convertOrder(data interface{}) *models.Order {
	o, ok := exchanges.BitfinexOrderToModel(data)
	if !ok {
		b.lg.Error(errors.Errorf("cant cast order to model %#v", data))
		return nil
	}
	return o
}

func (b *BitFinex) GetPositions() ([]*models.Position, error) {
	rs := make([]*models.Position, 0)

	ps := b.positions.GetAll()

	for _, p := range ps {
		rs = append(rs, b.convertPosition(p))
	}

	return rs, nil
}

func (b *BitFinex) convertPosition(data interface{}) *models.Position {
	o, ok := exchanges.BitfinexPositionToModel(data)
	if !ok {
		b.lg.Errorf("cant cast order to model %#v", data)
		return nil
	}
	return o
}

func (b *BitFinex) GetWallets() ([]*models.Wallets, error) {
	return []*models.Wallets{&b.walletsExchange, &b.walletsMargin}, nil
}

func (b *BitFinex) GetBalance() (models.BalanceUSD, error) {
	return b.balance, nil
}

// PutOrder https://docs.bitfinex.com/reference#ws-auth-input-order-new
func (b *BitFinex) PutOrder(o *models.PutOrder) (*models.Order, error) {
	if !b.ready {
		return nil, ErrNoConnect
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
			return nil, errors.WrapMessage(ErrInvalidRequestParams, "stop price are not specified")
		}
		Price = o.StopPrice
	case models.OrderTypeStopLimit:
		typeOrder = "STOP LIMIT"
		if o.Price == 0 {
			return nil, errors.WrapMessage(ErrInvalidRequestParams, "limit price are not specified")
		}
		PriceAuxLimit = o.Price
		if o.StopPrice == 0 {
			return nil, errors.WrapMessage(ErrInvalidRequestParams, "stop price are not specified")
		}
		Price = o.StopPrice
	default:
		return nil, ErrOrderTypeNotSupported
	}

	if !o.Margin {
		typeOrder = fmt.Sprintf("EXCHANGE %s", typeOrder)
	}

	if !strings.HasPrefix(Pair, "t") {
		return nil, ErrPairIncorrect
	}

	if o.InternalID != "" {
		orderClientID, err = strconv.ParseInt(o.InternalID, 10, 64)
		if err != nil {
			return nil, errors.WrapMessage(ErrInvalidRequestParams, err)
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
		AffiliateCode: b.cfg.Exchanges.BitFinex.Affiliate,
	}

	b.lg.Debugf("Submitting order: %#v", req)
	err = b.ws.SubmitOrder(b.ctx, req)
	if err != nil {
		return nil, err
	}

	orderPipe, err := b.watchers.New(fmt.Sprint("bf_wait_order", orderClientID), EventOrderFilled, EventOrderNew, EventRequestFail)
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
			case evt.Type == EventRequestFail:
				rr := evt.Payload.(RequestResult)
				if rr.Meta["order_id"] == o.InternalID {
					return nil, errors.WrapMessage(rr.Err, rr.Msg)
				}
			case evt.Type == EventOrderFilled, evt.Type == EventOrderNew:
				or := evt.Payload.(models.Order)
				if or.InternalID == o.InternalID {
					return &or, nil
				}
			}

		case <-Timout.C:
			return nil, ErrResultTimeOut
		}
	}
}

func (b *BitFinex) CancelOrder(o *models.Order) error {
	if !b.ready {
		return ErrNoConnect
	}

	id, err := strconv.ParseInt(o.ID, 10, 64)
	if err != nil {
		return errors.WrapMessage(ErrInvalidRequestParams, err)
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

	b.lg.Debugf("Canceling order: %#v", req)
	err = b.ws.SubmitCancel(b.ctx, &req)
	if err != nil {
		return err
	}

	orderPipe, err := b.watchers.New(fmt.Sprint("bf_wait_order", req.ID, req.CID), EventOrderCancel, EventRequestFail)
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
			case evt.Type == EventRequestFail:
				rr := evt.Payload.(RequestResult)
				if rr.Meta["order_id"] == o.ID || rr.Meta["order_id"] == o.InternalID {
					return errors.WrapMessage(rr.Err, rr.Msg)
				}
			case evt.Type == EventOrderCancel:
				or := evt.Payload.(models.Order)
				if or.ID == o.ID || or.InternalID == o.InternalID {
					return nil
				}
			}

		case <-Timout.C:
			return ErrResultTimeOut
		}
	}
}

// UpdateOrder if price, priceStop, amount equals 0, they will be ignored
func (b *BitFinex) UpdateOrder(orderID string, price float64, priceStop float64, amount float64) (*models.Order, error) {
	if !b.ready {
		return nil, ErrNoConnect
	}

	id, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, errors.WrapMessage(ErrInvalidRequestParams, err)
	}

	o := b.orders.Get(id)
	if o == nil {
		return nil, ErrOrderNotFound
	}

	req := &order.UpdateRequest{
		ID:     id,
		Price:  price,
		Amount: amount,
	}

	if priceStop != 0 {
		if !strings.Contains(o.Type, "STOP LIMIT") {
			return nil, errors.WrapMessage(ErrInvalidRequestParams, "order is not STOP LIMIT type")
		}

		req.PriceAuxLimit = price
		req.Price = priceStop
	}

	b.lg.Debugf("Updating order: %#v", req)
	err = b.ws.SubmitUpdateOrder(b.ctx, req)
	if err != nil {
		return nil, err
	}

	orderPipe, err := b.watchers.New(fmt.Sprint("bf_wait_order", id), EventOrderUpdate)
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
			case evt.Type == EventOrderUpdate:
				ord := evt.Payload.(models.Order)
				if ord.ID == orderID {
					return &ord, nil
				}
			}

		case <-Timout.C:
			return nil, ErrResultTimeOut
		}
	}
}

func (b *BitFinex) ClosePosition(p *models.Position) (*models.Position, error) {
	if !b.ready {
		return nil, ErrNoConnect
	}

	var Pair = p.Pair

	var prevStatePos = models.Position{}
	pos := b.positions.Get(p.GetIDAsInt())
	if pos == nil {
		return nil, ErrPositionNotFound
	}
	prevStatePos = *b.convertPosition(pos)

	req := &order.NewRequest{
		CID:           time.Now().Unix() / 1000,
		Symbol:        Pair,
		Type:          "MARKET",
		Amount:        -p.Amount,
		AffiliateCode: b.cfg.Exchanges.BitFinex.Affiliate,
		Close:         true,
	}

	b.lg.Debugf("Submitting order to close position: %#v", req)
	err := b.ws.SubmitOrder(b.ctx, req)
	if err != nil {
		return nil, err
	}

	positionPipe, err := b.watchers.New(fmt.Sprint("bf_wait_order", req.CID), EventPositionClosed, EventRequestFail)
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
			case evt.Type == EventRequestFail:
				rr := evt.Payload.(RequestResult)
				if rr.Meta["order_id"] == fmt.Sprint(req.CID) {
					return nil, errors.WrapMessage(rr.Err, rr.Msg)
				}
			case evt.Type == EventPositionClosed:
				pos := evt.Payload.(models.Position)
				if pos.ID == p.ID {
					return &prevStatePos, nil
				}
			}

		case <-Timout.C:
			return nil, ErrResultTimeOut
		}
	}
}

func (b *BitFinex) updateWallet(w *wallet.Wallet) *models.WalletCurrency {
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

func (b *BitFinex) CheckPair(pair string, margin bool) error {
	return checkPair(pair, margin)
}

func checkPair(pair string, margin bool) error {
	var pairs [][]string

	if !strings.HasPrefix(pair, "t") {
		return ErrPairIncorrect
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

	return ErrPairNotSupported
}
