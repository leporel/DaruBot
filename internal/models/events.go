package models

import (
	"DaruBot/pkg/watcher"
)

const (
	EventsModuleExchange watcher.ModuleType = "exchange"
)

var (
	EventTickerState = watcher.NewEventType(EventsModuleExchange, "EventTickerState", Ticker{})
	EventCandleState = watcher.NewEventType(EventsModuleExchange, "EventCandleState", Candle{})

	EventWalletUpdate = watcher.NewEventType(EventsModuleExchange, "EventWalletUpdate", WalletCurrency{})

	EventOrderPartiallyFilled = watcher.NewEventType(EventsModuleExchange, "EventOrderPartiallyFilled", Order{})
	EventOrderFilled          = watcher.NewEventType(EventsModuleExchange, "EventOrderFilled", Order{})
	EventOrderNew             = watcher.NewEventType(EventsModuleExchange, "EventOrderNew", Order{})
	EventOrderUpdate          = watcher.NewEventType(EventsModuleExchange, "EventOrderUpdate", Order{})
	EventOrderCancel          = watcher.NewEventType(EventsModuleExchange, "EventOrderCancel", Order{})

	EventPositionNew    = watcher.NewEventType(EventsModuleExchange, "EventPositionNew", Position{})
	EventPositionUpdate = watcher.NewEventType(EventsModuleExchange, "EventPositionUpdate", Position{})
	EventPositionClosed = watcher.NewEventType(EventsModuleExchange, "EventPositionClosed", Position{})

	EventError = watcher.NewEventType(EventsModuleExchange, "EventError", (*error)(nil))
)

type RequestResult struct {
	ReqID string
	Msg   string
	Err   error
	Meta  map[string]string
}
