package exchanges

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/watcher"
)

var (
	EventBookState = watcher.NewEvent(models.EventsModuleExchange, "EventBookState")
	EventPairState = watcher.NewEvent(models.EventsModuleExchange, "EventPairState")

	EventWalletUpdate = watcher.NewEvent(models.EventsModuleExchange, "EventWalletUpdate")

	EventOrderPartiallyFilled = watcher.NewEvent(models.EventsModuleExchange, "EventOrderPartiallyFilled")
	EventOrderFilled          = watcher.NewEvent(models.EventsModuleExchange, "EventOrderFilled")
	EventOrderNew             = watcher.NewEvent(models.EventsModuleExchange, "EventOrderNew")
	EventOrderUpdate          = watcher.NewEvent(models.EventsModuleExchange, "EventOrderUpdate")
	EventOrderCancel          = watcher.NewEvent(models.EventsModuleExchange, "EventOrderCancel")

	EventPositionNew    = watcher.NewEvent(models.EventsModuleExchange, "EventPositionNew")
	EventPositionUpdate = watcher.NewEvent(models.EventsModuleExchange, "EventPositionUpdate")
	EventPositionClosed = watcher.NewEvent(models.EventsModuleExchange, "EventPositionClosed")

	EventRequestSuccess = watcher.NewEvent(models.EventsModuleExchange, "EventRequestSuccess")
	EventRequestFail    = watcher.NewEvent(models.EventsModuleExchange, "EventRequestFail")

	EventError = watcher.NewEvent(models.EventsModuleExchange, "EventError")
)

type RequestResult struct {
	ReqID string
	Msg   string
	Err   error
	Meta  map[string]string
}
