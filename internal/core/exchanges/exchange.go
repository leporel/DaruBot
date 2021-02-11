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
	RegisterWatcher(name string, eType ...watcher.EventType) *watcher.Watcher
	RemoveWatcher(name string)
	EventsList() watcher.EventsMap

	/* Ticker */

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
	CancelOrder(order *models.Order) error
	ClosePosition(position *models.Position) (*models.Position, error)
}
