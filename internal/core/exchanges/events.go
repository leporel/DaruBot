package exchanges

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/watcher"
)

var (
	EventTickerState = watcher.NewEvent(models.EventsModuleExchange, "EventTickerState", models.Ticker{})
	EventCandleState = watcher.NewEvent(models.EventsModuleExchange, "EventCandleState", models.Candle{})

	EventWalletUpdate = watcher.NewEvent(models.EventsModuleExchange, "EventWalletUpdate", models.WalletCurrency{})

	EventOrderPartiallyFilled = watcher.NewEvent(models.EventsModuleExchange, "EventOrderPartiallyFilled", models.Order{})
	EventOrderFilled          = watcher.NewEvent(models.EventsModuleExchange, "EventOrderFilled", models.Order{})
	EventOrderNew             = watcher.NewEvent(models.EventsModuleExchange, "EventOrderNew", models.Order{})
	EventOrderUpdate          = watcher.NewEvent(models.EventsModuleExchange, "EventOrderUpdate", models.Order{})
	EventOrderCancel          = watcher.NewEvent(models.EventsModuleExchange, "EventOrderCancel", models.Order{})

	EventPositionNew    = watcher.NewEvent(models.EventsModuleExchange, "EventPositionNew", models.Position{})
	EventPositionUpdate = watcher.NewEvent(models.EventsModuleExchange, "EventPositionUpdate", models.Position{})
	EventPositionClosed = watcher.NewEvent(models.EventsModuleExchange, "EventPositionClosed", models.Position{})

	EventRequestSuccess = watcher.NewEvent(models.EventsModuleExchange, "EventRequestSuccess", models.Position{})
	EventRequestFail    = watcher.NewEvent(models.EventsModuleExchange, "EventRequestFail", models.Position{})

	EventError = watcher.NewEvent(models.EventsModuleExchange, "EventError", (*error)(nil))
)

type RequestResult struct {
	ReqID string
	Msg   string
	Err   error
	Meta  map[string]string
}
