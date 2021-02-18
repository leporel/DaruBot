package exchanges

type ExchangeType string

const (
	ExchangeTypeBitfinex ExchangeType = "Bitfinex"
)

func (e ExchangeType) String() string {
	return string(e)
}
