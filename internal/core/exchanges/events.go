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
	EventOrderCancel

	EventPositionNew
	EventPositionClosed

	EventRequestSuccess
	EventRequestFail

	EventError
)

type RequestResult struct {
	ID  string
	Msg string
	Err error
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
	case EventOrderCancel:
		return "EventOrderCancel"
	case EventPositionNew:
		return "EventPositionNew"
	case EventPositionClosed:
		return "EventPositionClosed"
	case EventRequestSuccess:
		return "EventRequestSuccess"
	case EventRequestFail:
		return "EventRequestFail"
	case EventError:
		return "EventError"
	default:
		return ""
	}
}
