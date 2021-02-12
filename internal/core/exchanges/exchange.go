package exchanges

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/watcher"
	"time"
)

type Exchange interface {
	/* Network */
	Connect() error
	Disconnect()
	IsReady() bool
	Ready() <-chan interface{}

	/* Watcher */
	SupportEvents() watcher.EventsMap

	/* Ticker */
	GetTicker(pair string) (*models.Ticker, error)
	GetCandles(pair string, resolution models.CandleResolution, start time.Time, end time.Time) (*models.Candles, error)

	GetSubscriptions() *models.Subscriptions
	SubscribeTicker(pair string) (subID string, err error)
	SubscribeCandles(pair string, resolution models.CandleResolution) (subID string, err error)
	Unsubscribe(subID string) error

	/* Tools */
	CheckPair(pair string, margin bool) error

	/* Data */
	GetOrders() ([]*models.Order, error)
	GetPositions() ([]*models.Position, error)
	GetWallets() ([]*models.Wallets, error)
	GetBalance() (models.BalanceUSD, error)
	HasUpdates(t time.Time) bool

	/* Requests */
	PutOrder(order *models.PutOrder) (*models.Order, error)
	UpdateOrder(orderID string, price float64, priceStop float64, amount float64) (*models.Order, error)
	CancelOrder(order *models.Order) error
	ClosePosition(position *models.Position) (*models.Position, error)
}
