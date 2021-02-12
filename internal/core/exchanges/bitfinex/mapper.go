package bitfinex

import (
	"DaruBot/internal/models"
	"DaruBot/internal/models/exchanges/bitfinex"
	"DaruBot/pkg/errors"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
)

func (b *Bitfinex) convertOrder(data interface{}) *models.Order {
	o, ok := bitfinex.BitfinexOrderToModel(data)
	if !ok {
		b.lg.Error(errors.Errorf("cant cast order to model %#v", data))
		return nil
	}
	return o
}

func (b *Bitfinex) convertPosition(data interface{}) *models.Position {
	o, ok := bitfinex.BitfinexPositionToModel(data)
	if !ok {
		b.lg.Errorf("cant cast position to model %#v", data)
		return nil
	}
	return o
}

func (b *Bitfinex) convertTicker(data interface{}) *models.Ticker {
	o, ok := bitfinex.BitfinexTickerToModel(data)
	if !ok {
		b.lg.Errorf("cant cast ticker to model %#v", data)
		return nil
	}
	return o
}

func (b *Bitfinex) convertCandle(data *candle.Candle) *models.Candle {
	o, ok := bitfinex.BitfinexCandleToModel(data)
	if !ok {
		b.lg.Errorf("cant cast candle to model %#v", data)
		return nil
	}
	return o
}
