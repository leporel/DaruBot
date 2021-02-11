package exchanges

import "DaruBot/pkg/errors"

var (
	ErrNoConnect     = errors.New("WEBSOCKET NOT CONNECTED")
	ErrResultTimeOut = errors.New("WEBSOCKET WAIT REQUEST RESULT TIMEOUT")

	ErrWebsocketError = errors.New("WEBSOCKET ERROR")

	ErrRequestError         = errors.New("REQUEST ERROR")
	ErrInvalidRequestParams = errors.New("INVALID REQUEST PARAMETERS")

	ErrOrderTypeNotSupported = errors.New("ORDER TYPE IS NOT SUPPORTED")
	ErrOrderNotFound         = errors.New("ORDER NOT FOUND")

	ErrPositionNotFound = errors.New("POSITION NOT FOUND")

	ErrPairIncorrect    = errors.New("PAIR INCORRECT")
	ErrPairNotSupported = errors.New("PAIR NOT SUPPORTED")
)
