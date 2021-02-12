package models

import (
	"DaruBot/pkg/watcher"
)

const (
	EventsModuleExchange watcher.ModuleType = iota
)

type ExchangeType string

const (
	ExchangeTypeBitfinex ExchangeType = "Bitfinex"
)
