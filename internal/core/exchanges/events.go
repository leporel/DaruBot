package exchanges

import "DaruBot/pkg/watcher"

const (
	EventBookState watcher.EventType = iota
	EventPairState

	EventOrderFilled
	EventOrderNew
	EventOrderUpdate

	EventPositionNew
	EventPositionClosed

	EventError
)
