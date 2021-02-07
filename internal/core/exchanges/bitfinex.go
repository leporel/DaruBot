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
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/order"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/position"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/ticker"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/trade"
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

	log := logging.MustGetLogger("bitfinex-ws")
	logging.SetLevel(logging.DEBUG, "bitfinex-ws")
	if !c.IsDebug() {
		logging.SetLevel(logging.ERROR, "bitfinex-ws")
	}
	logging.SetBackend(logger2.ConvertToGoLogging(lg.WithPrefix("exchange_ws", "Bitfinex")))
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
				b.lg.Debugf("InfoEvent: %#v", data)

				// TODO check PlatformInfo.Status https://docs.bitfinex.com/reference#rest-public-platform-status

			case *websocket.SubscribeEvent:
				// TODO handle

			case *websocket.UnsubscribeEvent:
				// TODO handle

			case *wallet.Snapshot:
				b.lg.Debugf("wallet snapshot %#v", data)

				for _, w := range data.Snapshot {
					b.updateWallet(w)
				}

			case *wallet.Update:
				b.lg.Debugf("wallet update %#v", data)

				w := wallet.Wallet(*data)
				b.updateWallet(&w)

			case *balanceinfo.Update:
				b.lg.Debugf("balance info %#v", data)

				b.balance.Total = data.TotalAUM
				b.balance.Total = data.NetAUM

			case *position.Snapshot:
				b.lg.Debugf("position snapshot %#v", data)

				for _, p := range data.Snapshot {
					b.positions.Add(p)
				}

			case *position.Update:
				b.lg.Debugf("position update %#v", data)

				if data.Status == "CLOSED" {
					b.positions.Delete(data.Id)
				} else {
					p := position.Position(*data)
					b.positions.Add(&p)
				}

			case *position.New:
				b.lg.Debugf("position new %#v", data)
				p := position.Position(*data)
				b.positions.Add(&p)

			case *position.Cancel:
				b.lg.Debugf("position cancel %#v", data)

				b.positions.Delete(data.Id)

			case *order.Snapshot:
				b.lg.Debugf("order snapshot %#v", data)

				for _, p := range data.Snapshot {
					b.orders.Add(p)
				}

			case *order.Update:
				b.lg.Debugf("order update %#v", data)

				if strings.Contains(data.Status, "EXECUTED") {
					b.orders.Delete(data.ID)
				}
				if strings.Contains(data.Status, "PARTIALLY FILLED") {
					b.lg.Infof("order %v was partially filled", data.ID)
				}

			case *order.Cancel:
				b.lg.Debugf("order cancel %#v", data)

				b.orders.Delete(data.ID)

			case *order.New:
				b.lg.Debugf("new order created %#v", data)

				o := order.Order(*data)
				b.orders.Add(&o)

			case *trade.Trade:
				b.lg.Debugf("new trade: %s", data)

			case *ticker.Snapshot:
				b.lg.Debugf("new ticker: %s", data)

				// TODO handle

			case *ticker.Update:
				b.lg.Debugf("ticker update: %s", data)

				// TODO handle

			case error:
				b.lg.Errorf("channel closed: %s", data)
				b.watchers.Emmit(watcher.NewEvent(EventError, errors.New(fmt.Sprintf("channel closed: %s", data))))
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

func (b *BitFinex) GetOrders() ([]Order, error) {
	rs := make([]Order, 0)

	orders := b.orders.GetAll()

	for _, o := range orders {
		rs = append(rs, exchanges.BitfinexOrderToModel(o))
	}

	return rs, nil
}

func (b *BitFinex) GetPositions() ([]Position, error) {
	rs := make([]Position, 0)

	ps := b.positions.GetAll()

	for _, p := range ps {
		rs = append(rs, exchanges.BitfinexPositionToModel(p))
	}

	return rs, nil
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
	case models.OrderTypeStop:
		typeOrder = "STOP"
	case models.OrderTypeStopLimit:
		typeOrder = "STOP LIMIT"
		PriceAuxLimit = o.Price
		Price = o.StopPrice
	default:
		return OrderTypeNotSupported
	}

	if !o.Margin {
		typeOrder = fmt.Sprintf("EXCHANGE %s", typeOrder)
	}

	if o.MarketPrice {
		Amount = 0
	}

	if !strings.HasPrefix(Pair, "t") {
		return PairIncorrect
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
}

func (b *BitFinex) CheckPair(pair string, margin bool) error {
	return checkPair(pair, margin)
}

func checkPair(pair string, margin bool) error {
	var pairs [][]string

	if !strings.HasPrefix(pair, "t") {
		return PairIncorrect
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

	return PairNotSupported
}
