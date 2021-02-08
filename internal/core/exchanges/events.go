package exchanges

import (
	"DaruBot/pkg/watcher"
)

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

func EventToString(t watcher.EventType) string {
	switch t {
	case EventBookState:
		return "EventBookState"
	case EventPairState:
		return "EventPairState"
	case EventWalletUpdate:
		return "EventWalletUpdate"
	case EventOrderPartiallyFilled:
		return "EventOrderPartiallyFilled"
	case EventOrderFilled:
		return "EventOrderFilled"
	case EventOrderNew:
		return "EventOrderNew"
	case EventOrderUpdate:
		return "EventOrderUpdate"
	case EventPositionNew:
		return "EventPositionNew"
	case EventPositionClosed:
		return "EventPositionClosed"
	case EventRequestSuccess:
		return "EventRequestSuccess"
	case EventError:
		return "EventError"
	default:
		return ""
	}
}
