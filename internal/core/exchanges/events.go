package exchanges

import "DaruBot/pkg/watcher"

const (
	EventBookState watcher.EventType = iota
	EventPairState

	EventWalletUpdate

	EventOrderPartiallyFilled
	EventOrderFilled
	EventOrderNew
	EventOrderUpdate

	EventPositionNew
	EventPositionClosed

	EventRequestSuccess

	EventError
)

type RequestResult struct {
	Msg string
}
