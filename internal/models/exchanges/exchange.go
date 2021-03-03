package exchanges

type ExchangeType string

const (
	ExchangeTypeMock     ExchangeType = "CryptoMock"
	ExchangeTypeBitfinex ExchangeType = "Bitfinex"
)

func (e ExchangeType) String() string {
	return string(e)
}
