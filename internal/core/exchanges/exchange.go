package exchanges

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/watcher"
)

type Exchange interface {
	Connect() error
	Disconnect()
	IsReady() bool
	Ready() <-chan interface{}

	RegisterWatcher(name string, eType ...watcher.EventType) *watcher.Watcher
	RemoveWatcher(name string)
	EventsList() []watcher.EventType

	CheckPair(pair string, margin bool) error

	GetOrders() ([]Order, error)
	GetPositions() ([]Position, error)
	GetWallets() ([]*models.Wallets, error)
	GetBalance() (models.BalanceUSD, error)

	PutOrder(order *models.PutOrder) (*models.Order, error)
	CancelOrder(order *models.Order) error
	ClosePosition(position *models.Position) error
}

type Order interface {
	GetID() string
	GetPrice() float64
	GetAmount() float64
	GetOriginalAmount() float64
	GetType() models.OrderType
}

type Position interface {
	GetID() string
	GetPrice() float64
	GetAmount() float64
	GetLiquidationPrice() float64
	GetMarginLevel() float64
	GetProfit() float64
	GetProfitPercentage() float64
}
