package exchanges

import (
	"DaruBot/pkg/errors"
)

var (
	ErrNoConnect     = errors.New("WEBSOCKET NOT CONNECTED")
	ErrNotOperate    = errors.New("PLATFORM ARE NOT OPERATE")
	ErrResultTimeOut = errors.New("WEBSOCKET WAIT REQUEST RESULT TIMEOUT")

	ErrWebsocketError = errors.New("WEBSOCKET ERROR")

	ErrRequestError         = errors.New("REQUEST ERROR")
	ErrInvalidRequestParams = errors.New("INVALID REQUEST PARAMETERS")

	ErrOrderTypeNotSupported = errors.New("ORDER TYPE IS NOT SUPPORTED")
	ErrOrderNotFound         = errors.New("ORDER NOT FOUND")

	ErrPositionNotFound = errors.New("POSITION NOT FOUND")

	ErrSymbolIncorrect    = errors.New("PAIR INCORRECT")
	ErrSymbolNotSupported = errors.New("PAIR NOT SUPPORTED")
)
