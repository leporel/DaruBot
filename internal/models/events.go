package models

import (
	"DaruBot/pkg/watcher"
)

const (
	EventsModuleExchange watcher.ModuleType = "exchange"
)

var (
	EventTickerState = watcher.NewEvent(EventsModuleExchange, "EventTickerState", Ticker{})
	EventCandleState = watcher.NewEvent(EventsModuleExchange, "EventCandleState", Candle{})

	EventWalletUpdate = watcher.NewEvent(EventsModuleExchange, "EventWalletUpdate", WalletCurrency{})

	EventOrderPartiallyFilled = watcher.NewEvent(EventsModuleExchange, "EventOrderPartiallyFilled", Order{})
	EventOrderFilled          = watcher.NewEvent(EventsModuleExchange, "EventOrderFilled", Order{})
	EventOrderNew             = watcher.NewEvent(EventsModuleExchange, "EventOrderNew", Order{})
	EventOrderUpdate          = watcher.NewEvent(EventsModuleExchange, "EventOrderUpdate", Order{})
	EventOrderCancel          = watcher.NewEvent(EventsModuleExchange, "EventOrderCancel", Order{})

	EventPositionNew    = watcher.NewEvent(EventsModuleExchange, "EventPositionNew", Position{})
	EventPositionUpdate = watcher.NewEvent(EventsModuleExchange, "EventPositionUpdate", Position{})
	EventPositionClosed = watcher.NewEvent(EventsModuleExchange, "EventPositionClosed", Position{})

	EventRequestSuccess = watcher.NewEvent(EventsModuleExchange, "EventRequestSuccess", Position{})
	EventRequestFail    = watcher.NewEvent(EventsModuleExchange, "EventRequestFail", Position{})

	EventError = watcher.NewEvent(EventsModuleExchange, "EventError", (*error)(nil))
)

type RequestResult struct {
	ReqID string
	Msg   string
	Err   error
	Meta  map[string]string
}
