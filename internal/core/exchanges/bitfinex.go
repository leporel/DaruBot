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
	"strings"
	"time"
)

var (
	supportEvents = []watcher.EventType{
		EventError,
		EventRequestSuccess,
		EventOrderNew,
		EventOrderFilled,
		EventOrderPartiallyFilled,
		EventWalletUpdate,
	}
)

type BitFinex struct {
	ctx context.Context

	ws   *websocket.Client
	rest *rest.Client

	lg logger.Logger

	ready          bool
	readyChan      chan interface{}
	disconnectChan chan interface{}

	orders    *exchanges.BitfinexOrders
	positions *exchanges.BitfinexPositions

	walletsExchange models.Wallets
	walletsMargin   models.Wallets
	balance         models.BalanceUSD

	watchers *watcher.WatcherManager
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
		watchers:        watcher.NewWatcherManager(),
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

	errorPipe := b.watchers.New("_api_Errors", EventError)
	defer b.watchers.Remove("_api_Errors")

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

func (b *BitFinex) EventsList() []watcher.EventType {
	return supportEvents
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

				for _, w := range data.Snapshot {
					b.updateWallet(w)
				}

			case *wallet.Update:
				b.lg.Debugf("WALLET UPDATE %#v", data)

				w := wallet.Wallet(*data)
				b.updateWallet(&w)

			case *balanceinfo.Update:
				b.lg.Debugf("BALANCE INFO %#v", data)

				b.balance.Total = data.TotalAUM
				b.balance.Total = data.NetAUM

			case *position.Snapshot:
				b.lg.Debugf("POSITION SNAPSHOT %#v", data)

				for _, p := range data.Snapshot {
					b.positions.Add(p)
				}

			case *position.Update:
				b.lg.Debugf("POSITION UPDATE %#v", data)

				if data.Status == "CLOSED" {
					b.positions.Delete(data.Id)
				} else {
					p := position.Position(*data)
					b.positions.Add(&p)
				}

			case *position.New:
				b.lg.Debugf("POSITION NEW %#v", data)
				p := position.Position(*data)
				b.positions.Add(&p)

			case *position.Cancel:
				b.lg.Debugf("POSITION CANCEL %#v", data)

				b.positions.Delete(data.Id)

			case *order.Snapshot:
				b.lg.Debugf("ORDER SNAPSHOT %#v", data)

				for _, p := range data.Snapshot {
					b.orders.Add(p)
				}

			case *order.Update:
				b.lg.Debugf("ORDER UPDATE %#v", data)

				if strings.Contains(data.Status, "PARTIALLY FILLED") {
					//b.lg.Infof("order %v was partially filled at %v, left amount %v", data.ID, data.Price, data.Amount)
					b.emmit(EventOrderPartiallyFilled, *b.convertOrder(data))
				}

			case *order.Cancel:
				b.lg.Debugf("ORDER CANCEL %#v", data)

				if strings.Contains(data.Status, "EXECUTED") {
					b.emmit(EventOrderFilled, *b.convertOrder(data))
				}

				b.orders.Delete(data.ID)

			case *order.New:
				b.lg.Debugf("ORDER NEW %#v", data)

				if data.Status == "ACTIVE" {
					b.emmit(EventOrderNew, *b.convertOrder(data))
				}

				o := order.Order(*data)
				b.orders.Add(&o)

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
				switch data.Status {
				case "ERROR", "FAILURE":
					b.emmit(EventError, errors.WrapMessage(ErrRequestError, data.Text))
				case "SUCCESS":
					b.emmit(EventRequestSuccess, RequestResult{Msg: data.Text})
				default:
					b.lg.Warn("UNKNOWN NOTIFICATION:  %#v", data)
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
	return b.watchers.New(name, eType...)
}

func (b *BitFinex) RemoveWatcher(name string) {
	b.watchers.Remove(name)
}

func (b *BitFinex) emmit(Event watcher.EventType, data interface{}) {
	b.watchers.Emmit(watcher.NewEvent(Event, data))
}

func (b *BitFinex) GetOrders() ([]Order, error) {
	rs := make([]Order, 0)

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

func (b *BitFinex) GetPositions() ([]Position, error) {
	rs := make([]Position, 0)

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
func (b *BitFinex) PutOrder(o *models.PutOrder) error {
	var typeOrder string
	var PriceAuxLimit float64
	var Price = o.Price
	var Amount = o.Amount
	var Pair = o.Pair

	switch o.Type {
	case models.OrderTypeLimit:
		typeOrder = "LIMIT"
	case models.OrderTypeMarket:
		typeOrder = "MARKET"
		Price = 0
	case models.OrderTypeStop:
		typeOrder = "STOP"
		if o.StopPrice == 0 {
			return errors.WrapMessage(ErrOrderBadPutOrderParams, "stop price are not specified")
		}
		Price = o.StopPrice
	case models.OrderTypeStopLimit:
		typeOrder = "STOP LIMIT"
		if o.Price == 0 {
			return errors.WrapMessage(ErrOrderBadPutOrderParams, "limit price are not specified")
		}
		PriceAuxLimit = o.Price
		if o.StopPrice == 0 {
			return errors.WrapMessage(ErrOrderBadPutOrderParams, "stop price are not specified")
		}
		Price = o.StopPrice
	default:
		return ErrOrderTypeNotSupported
	}

	if !o.Margin {
		typeOrder = fmt.Sprintf("EXCHANGE %s", typeOrder)
	}

	if !strings.HasPrefix(Pair, "t") {
		return ErrPairIncorrect
	}

	req := &order.NewRequest{
		GID:           0,
		CID:           time.Now().Unix() / 1000,
		Type:          typeOrder,
		Symbol:        Pair,
		Amount:        Amount,
		Price:         Price,
		PriceAuxLimit: PriceAuxLimit,
		AffiliateCode: "", // TODO https://www.bitfinex.com/affiliates/
	}

	b.lg.Debugf("Submitting order: %#v", req)
	err := b.ws.SubmitOrder(b.ctx, req)
	if err != nil {
		return err
	}

	return nil
}

//func (b *BitFinex) CancelOrder(order *models.PutOrder) error

//func (b *BitFinex) ClosePosition(position *models.Position) error {
//  use https://docs.bitfinex.com/docs/flag-values
//}

func (b *BitFinex) updateWallet(w *wallet.Wallet) {
	wl := &models.WalletCurrency{
		Name:      w.Currency,
		Available: w.BalanceAvailable,
		Balance:   w.Balance,
	}

	switch w.Type {
	case "exchange":
		b.walletsExchange.Add(wl)
	case "margin":
		b.walletsMargin.Add(wl)
	}

	b.emmit(EventWalletUpdate, *wl)
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
