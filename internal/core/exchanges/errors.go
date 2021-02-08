package exchanges

import "DaruBot/pkg/errors"

var (
	ErrWebsocketError = errors.New("WEBSOCKET ERROR")
	ErrRequestError   = errors.New("REQUEST ERROR")

	ErrOrderTypeNotSupported  = errors.New("ORDER TYPE IS NOT SUPPORTED")
	ErrOrderBadPutOrderParams = errors.New("PUT ORDER BAD PARAMETERS")

	ErrPairIncorrect    = errors.New("PAIR INCORRECT")
	ErrPairNotSupported = errors.New("PAIR NOT SUPPORTED")
)
