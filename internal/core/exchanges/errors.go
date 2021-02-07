package exchanges

import "DaruBot/pkg/errors"

var (
	OrderTypeNotSupported = errors.New("ORDER TYPE IS NOT SUPPORTED")

	PairIncorrect    = errors.New("PAIR INCORRECT")
	PairNotSupported = errors.New("PAIR NOT SUPPORTED")
)
