package exchanges

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/watcher"
)

var (
	EventTickerState = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventTickerState", models.Ticker{})
	EventCandleState = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventCandleState", models.Candle{})

	EventWalletUpdate = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventWalletUpdate", models.WalletCurrency{})

	EventOrderPartiallyFilled = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventOrderPartiallyFilled", models.Order{})
	EventOrderFilled          = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventOrderFilled", models.Order{})
	EventOrderNew             = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventOrderNew", models.Order{})
	EventOrderUpdate          = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventOrderUpdate", models.Order{})
	EventOrderCancel          = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventOrderCancel", models.Order{})

	EventPositionNew    = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventPositionNew", models.Position{})
	EventPositionUpdate = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventPositionUpdate", models.Position{})
	EventPositionClosed = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventPositionClosed", models.Position{})

	EventRequestSuccess = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventRequestSuccess", models.Position{})
	EventRequestFail    = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventRequestFail", models.Position{})

	EventError = watcher.NewEvent(watcher.ModuleType(models.EventsModuleExchange), "EventError", (*error)(nil))
)

type RequestResult struct {
	ReqID string
	Msg   string
	Err   error
	Meta  map[string]string
}
