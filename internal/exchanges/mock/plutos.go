package mock

import (
	"DaruBot/internal/models"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"sync"
	"time"
)

type TickerFunc func(symbol string, curTime time.Time) (*models.Ticker, error)

type Plutos struct {
	SubscribeManager *subscribeManager
	mu               *sync.Mutex
	work             bool
	channel          chan interface{}
	getTicker        TickerFunc

	maxLeverage uint8
	taxFee      float64 // TODO implement
	wallets     *models.Wallets
	positions   []models.Position
	orders      []models.Order

	currentTime time.Time

	// Fiat money name (e.g. USDT)
	currency string
}

var (
	ErrNotExecuted = fmt.Errorf("order not executed")
)

func NewPlutos(maxLeverage uint8, taxFee float64, currency string, w *models.Wallets, o []models.Order, p []models.Position) *Plutos {
	sm := &subscribeManager{
		subs: []*subscription{},
	}

	return &Plutos{
		SubscribeManager: sm,
		mu:               &sync.Mutex{},
		work:             true,
		channel:          make(chan interface{}, 100),

		maxLeverage: maxLeverage,
		wallets:     w,
		positions:   p,
		taxFee:      taxFee,
		orders:      o,
		currency:    currency,
	}
}

func (p *Plutos) SetTickerFunc(tf TickerFunc) {
	p.getTicker = tf
}

func (p *Plutos) GetChan() chan interface{} {
	return p.channel
}

func (p *Plutos) Listen(ch <-chan time.Time) {
	for {
		select {
		case t := <-ch:
			p.mu.Lock()
			p.SubscribeManager.trigger(t, p.channel)
			p.currentTime = t

			if checkResTiming(1*time.Minute, t) {
				if len(p.orders) > 0 {
					if err := p.processOrders(); err != nil {
						panic(err)
					}
				}
				if len(p.positions) > 0 {
					if err := p.processPositions(); err != nil {
						panic(err)
					}
				}

			}
			p.mu.Unlock()
		default:
			if p.work == false {
				return
			}
		}
	}
}

func (p *Plutos) Stop() {
	p.work = false
}

func (p *Plutos) GetOrders() []*models.Order {
	p.mu.Lock()
	defer p.mu.Unlock()
	rs := make([]*models.Order, 0)
	for _, order := range p.orders {
		o := order
		rs = append(rs, &o)
	}
	return rs
}

func (p *Plutos) GetPositions() []*models.Position {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.getPositions()
}

func (p *Plutos) getPositions() []*models.Position {
	rs := make([]*models.Position, 0)
	for _, position := range p.positions {
		pos := position
		rs = append(rs, &pos)
	}
	return rs
}

func (p *Plutos) GetWallets() []*models.Wallets {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.getWallets()
}

func (p *Plutos) getWallets() []*models.Wallets {
	rs := make([]*models.Wallets, 0)

	w := models.Wallets{
		WalletType: p.wallets.WalletType,
	}

	for _, wallet := range p.wallets.GetAll() {
		cur := *wallet
		w.Update(&cur)
	}

	rs = append(rs, &w)

	return rs
}

func (p *Plutos) GetBalance(curTime time.Time) (*models.BalanceUSD, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	rs := &models.BalanceUSD{
		Total:    0,
		NetWorth: 0,
	}

	ws := p.getWallets()

	mapCurs := make(map[string]float64)

	for _, w := range ws {
		for _, currency := range w.GetAll() {
			if currency.Name == p.currency {
				rs.Total = currency.Balance
				rs.NetWorth = currency.Balance
				continue
			}
			mapCurs[currency.Name] = currency.Balance
		}
	}

	pos := p.getPositions()

	for _, po := range pos {
		ticker, err := p.getTicker(fmt.Sprintf("%s", po.Symbol), curTime)
		if err != nil {
			return nil, err
		}
		rs.NetWorth = rs.NetWorth + po.Amount*ticker.Price
	}

	for name, value := range mapCurs {
		ticker, err := p.getTicker(fmt.Sprintf("%s%s", name, p.currency), curTime)
		if err != nil {
			return nil, err
		}
		rs.NetWorth = rs.NetWorth + value*ticker.Price
	}

	return rs, nil
}

func (p *Plutos) PutOrder(putOrder *models.PutOrder) (models.Order, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if putOrder.Amount == 0 {
		return models.Order{}, fmt.Errorf("wrong amount")
	}
	if putOrder.Price == 0 && putOrder.Type != models.OrderTypeMarket {
		return models.Order{}, fmt.Errorf("wrong price")
	}

	o := models.Order{
		ID:             uuid.Must(uuid.NewUUID()).String(),
		Symbol:         putOrder.Symbol,
		InternalID:     putOrder.InternalID,
		Type:           putOrder.Type,
		Price:          putOrder.Price,
		PriceAvg:       0,
		AmountCurrent:  putOrder.Amount,
		AmountOriginal: putOrder.Amount,
		Date:           p.currentTime,
		Updated:        p.currentTime,
		Meta:           map[string]interface{}{"stop_price": putOrder.StopPrice},
	}

	ticker, err := p.getTicker(putOrder.Symbol, p.currentTime)
	if err != nil {
		return models.Order{}, err
	}
	executedCost := ticker.Price * o.AmountOriginal

	walletAsset, walletCurrency := p.relatedWallets(putOrder.Symbol)
	sell := o.IsSellOrder()
	if sell && walletAsset.Available < o.AmountOriginal {
		return models.Order{}, fmt.Errorf("insufficient %s balance", walletAsset.Name)
	} else if walletCurrency.Available < executedCost {
		return models.Order{}, fmt.Errorf("insufficient %s balance", walletCurrency.Name)
	}

	switch o.Type {
	case models.OrderTypeMarket:
		rs, err := p.executeOrder(&o, ticker)
		if err != nil {
			return models.Order{}, err
		}
		return *rs, nil
	case models.OrderTypeStop, models.OrderTypeLimit:
		p.orders = append(p.orders, o)

		if sell {
			walletAsset.Available = walletAsset.Available - o.AmountOriginal
			p.walletEvent(*walletAsset)
		} else {
			walletCurrency.Available = walletCurrency.Available - executedCost
			p.walletEvent(*walletCurrency)
		}
		p.wallets.Update(walletAsset)
		p.wallets.Update(walletCurrency)

		p.orderEvent(o)

	default:
		return models.Order{}, fmt.Errorf("order type not supported")
	}

	return o, fmt.Errorf("order not processed")
}

func (p *Plutos) processOrders() error {
	tickers := make(map[string]*models.Ticker, 0)
	var err error

	for i, order := range p.orders {
		var executed, ok bool
		var ticker *models.Ticker

		if ticker, ok = tickers[order.Symbol]; !ok {
			ticker, err = p.getTicker(order.Symbol, p.currentTime)
			if err != nil {
				return err
			}
			tickers[order.Symbol] = ticker
		}

		if _, err = p.executeOrder(&order, ticker); err != nil {
			if err == ErrNotExecuted {
				err = nil
			} else {
				return err
			}
		} else {
			executed = true
		}

		if executed {
			p.orders = append(p.orders[:i], p.orders[i+1:]...)
		}
	}

	return nil
}

func (p *Plutos) executeOrder(order *models.Order, ticker *models.Ticker) (*models.Order, error) {
	sell := order.IsSellOrder()

	switch order.Type {
	case models.OrderTypeMarket:
		return p.orderApply(order, ticker, sell)

	case models.OrderTypeLimit:
		if ticker.Price >= order.Price && sell {
			return p.orderApply(order, ticker, sell)
		}
		if ticker.Price <= order.Price && !sell {
			return p.orderApply(order, ticker, sell)
		}
	case models.OrderTypeStop:
		val, ok := order.Meta["stop_price"]
		if !ok {
			return nil, fmt.Errorf("stop price are not specified")
		}
		stopPrice, _ := val.(float64)
		if ticker.Price <= stopPrice && sell {
			return p.orderApply(order, ticker, sell)
		}
		if ticker.Price >= stopPrice && !sell {
			return p.orderApply(order, ticker, sell)
		}
	default:
		return nil, fmt.Errorf("order type not supported")
	}

	return nil, ErrNotExecuted
}

func (p *Plutos) orderApply(order *models.Order, ticker *models.Ticker, sell bool) (*models.Order, error) {
	order.AmountCurrent = 0
	order.Updated = p.currentTime
	order.PriceAvg = ticker.Price

	walletAsset, walletCurrency := p.relatedWallets(order.Symbol)

	executedCost := order.AmountOriginal * ticker.Price

	// TODO if margin make position

	if sell {
		walletAsset.Balance = walletAsset.Balance - order.AmountOriginal
		walletCurrency.Balance = walletCurrency.Balance + executedCost
		walletCurrency.Available = walletCurrency.Available + executedCost
	} else {
		walletAsset.Balance = walletAsset.Balance + order.AmountOriginal
		walletAsset.Available = walletAsset.Available + order.AmountOriginal
		walletCurrency.Balance = walletCurrency.Balance - executedCost
	}

	p.walletEvent(*walletAsset)
	p.walletEvent(*walletCurrency)
	p.orderEvent(*order)

	return order, nil
}

func (p *Plutos) orderEvent(order models.Order) {
	p.channel <- &order
}

func (p *Plutos) walletEvent(wc models.WalletCurrency) {
	p.channel <- &wc
}

func (p *Plutos) relatedWallets(pair string) (*models.WalletCurrency, *models.WalletCurrency) {
	asset := strings.TrimSuffix(pair, p.currency)

	walletAsset := p.wallets.Get(asset)
	if walletAsset == nil {
		walletAsset = &models.WalletCurrency{
			Name:       asset,
			WalletType: models.WalletTypeNone,
			Balance:    0,
			Available:  0,
		}
	}
	walletCurrency := p.wallets.Get(p.currency)
	if walletCurrency == nil {
		walletCurrency = &models.WalletCurrency{
			Name:       p.currency,
			WalletType: models.WalletTypeNone,
			Balance:    0,
			Available:  0,
		}
	}

	return walletAsset, walletCurrency
}

func (p *Plutos) addPosition() {

}

func (p *Plutos) processPositions() error {

	return nil
}

func (p *Plutos) positionClose() {

}
