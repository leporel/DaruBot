package bitfinex

import (
	"DaruBot/internal/models"
	"DaruBot/internal/models/exchanges"
	"DaruBot/pkg/errors"
	"DaruBot/pkg/tools"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/candle"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/order"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/position"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/ticker"
)

func (b *bitfinexWebsocket) convertOrder(data interface{}) *models.Order {
	o, ok := bitfinexOrderToModel(data)
	if !ok {
		b.log.Error(errors.Errorf("cant cast order to model %#v", data))
		return nil
	}
	return o
}

func (b *bitfinexWebsocket) convertPosition(data interface{}) *models.Position {
	o, ok := bitfinexPositionToModel(data)
	if !ok {
		b.log.Errorf("cant cast position to model %#v", data)
		return nil
	}
	return o
}

func (b *bitfinexWebsocket) convertTicker(data interface{}) *models.Ticker {
	o, ok := bitfinexTickerToModel(data)
	if !ok {
		b.log.Errorf("cant cast ticker to model %#v", data)
		return nil
	}
	return o
}

func (b *bitfinexWebsocket) convertCandle(data *candle.Candle) *models.Candle {
	o, ok := bitfinexCandleToModel(data)
	if !ok {
		b.log.Errorf("cant cast candle to model %#v", data)
		return nil
	}
	return o
}

func bitfinexOrderToModel(or interface{}) (*models.Order, bool) {
	var o order.Order

	switch t := or.(type) {
	case *order.Order:
		o = *t
	case *order.Update:
		o = order.Order(*t)
	case *order.Cancel:
		o = order.Order(*t)
	case *order.New:
		o = order.Order(*t)
	default:
		return nil, false
	}

	rs := &models.Order{
		ID:             fmt.Sprint(o.ID),
		InternalID:     fmt.Sprint(o.CID),
		Symbol:         o.Symbol,
		Price:          o.Price,
		PriceAvg:       o.PriceAvg,
		AmountCurrent:  o.Amount,
		AmountOriginal: o.AmountOrig,
		Date:           tools.TimeFromMilliseconds(o.MTSCreated),
		Updated:        tools.TimeFromMilliseconds(o.MTSUpdated),
		Meta:           make(map[string]interface{}),
	}

	switch o.Type {
	case "LIMIT", "EXCHANGE LIMIT":
		rs.Type = models.OrderTypeLimit
	case "MARKET", "EXCHANGE MARKET":
		rs.Type = models.OrderTypeMarket
	case "STOP", "EXCHANGE STOP":
		rs.Type = models.OrderTypeStop
	case "STOP LIMIT", "EXCHANGE STOP LIMIT":
		rs.Type = models.OrderTypeStopLimit
	//case "TRAILING STOP", "EXCHANGE TRAILING STOP":
	//	rs.Type = models.OrderTypeTrailingStop
	default:
		rs.Type = models.OrderTypeUnknown
	}

	rs.Meta["Exchange"] = exchanges.ExchangeTypeBitfinex
	rs.Meta["Status"] = o.Status
	rs.Meta["Type"] = o.Type
	rs.Meta["Flags"] = o.Flags

	return rs, true
}

func bitfinexPositionToModel(po interface{}) (*models.Position, bool) {
	var p position.Position

	switch t := po.(type) {
	case *position.Position:
		p = *t
	case *position.Update:
		p = position.Position(*t)
	case *position.Cancel:
		p = position.Position(*t)
	case *position.New:
		p = position.Position(*t)
	default:
		return nil, false
	}

	rs := &models.Position{
		ID:                   fmt.Sprint(p.Id),
		Symbol:               p.Symbol,
		Price:                p.BasePrice,
		Amount:               p.Amount,
		LiqPrice:             p.LiquidationPrice,
		MarginLevel:          p.Leverage,
		ProfitLoss:           p.ProfitLoss,
		ProfitLossPercentage: p.ProfitLossPercentage,
		Meta:                 make(map[string]interface{}),
	}

	rs.Meta["Exchange"] = exchanges.ExchangeTypeBitfinex
	rs.Meta["Status"] = p.Status

	return rs, true
}

func bitfinexTickerToModel(tk interface{}) (*models.Ticker, bool) {
	var t ticker.Ticker

	switch tp := tk.(type) {
	case *ticker.Ticker:
		t = *tp
	case *ticker.Update:
		t = ticker.Ticker(*tp)
	default:
		return nil, false
	}

	rs := &models.Ticker{
		Pair:     t.Symbol,
		Price:    t.LastPrice,
		Exchange: exchanges.ExchangeTypeBitfinex,
		State: models.TickerState{
			High:    t.High,
			Low:     t.Low,
			Volume:  t.Volume,
			BID:     t.Bid,
			BIDSize: t.BidSize,
			ASK:     t.Ask,
			ASKSize: t.AskSize,
		},
	}

	return rs, true
}

func bitfinexCandleToModel(c *candle.Candle) (*models.Candle, bool) {

	rs := &models.Candle{
		Symbol:     c.Symbol,
		Resolution: candleBitfinexResolutionToModel(c.Resolution),
		Date:       tools.TimeFromMilliseconds(c.MTS),
		Open:       c.Open,
		Close:      c.Close,
		High:       c.High,
		Low:        c.Low,
		Volume:     c.Volume,
	}

	return rs, true
}

func candleResolutionToBitfinex(c models.CandleResolution) (common.CandleResolution, error) {
	switch c {
	case models.OneMinute:
		return common.OneMinute, nil
	case models.FiveMinutes:
		return common.FiveMinutes, nil
	case models.FifteenMinutes:
		return common.FifteenMinutes, nil
	case models.ThirtyMinutes:
		return common.ThirtyMinutes, nil
	case models.OneHour:
		return common.OneHour, nil
	case models.ThreeHours:
		return common.ThreeHours, nil
	case models.SixHours:
		return common.SixHours, nil
	case models.TwelveHours:
		return common.TwelveHours, nil
	case models.OneDay:
		return common.OneDay, nil
	case models.OneWeek:
		return common.OneWeek, nil
	case models.OneMonth:
		return common.OneMonth, nil
	default:
		return common.OneMinute, fmt.Errorf("could not convert string to resolution: %s", c)
	}
}

func candleBitfinexResolutionToModel(c common.CandleResolution) models.CandleResolution {
	switch c {
	case common.OneMinute:
		return models.OneMinute
	case common.FiveMinutes:
		return models.FiveMinutes
	case common.FifteenMinutes:
		return models.FifteenMinutes
	case common.ThirtyMinutes:
		return models.ThirtyMinutes
	case common.OneHour:
		return models.OneHour
	case common.ThreeHours:
		return models.ThreeHours
	case common.SixHours:
		return models.SixHours
	case common.TwelveHours:
		return models.TwelveHours
	case common.OneDay:
		return models.OneDay
	case common.OneWeek:
		return models.OneWeek
	case common.OneMonth:
		return models.OneMonth
	default:
		panic(fmt.Errorf("could not convert string to resolution: %s", c))
	}
}
