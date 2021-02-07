package exchanges

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/tools"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/order"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/position"
	"sync"
)

type BitfinexOrders struct {
	orders sync.Map
}

func (o *BitfinexOrders) Add(rd *order.Order) {
	o.orders.Store(rd.ID, rd)
}

func (o *BitfinexOrders) Get(orderID int64) *order.Order {
	rd, ok := o.orders.Load(orderID)

	if !ok {
		return nil
	}

	return rd.(*order.Order)
}

func (o *BitfinexOrders) Delete(orderID int64) *order.Order {
	rd, ok := o.orders.LoadAndDelete(orderID)

	if !ok {
		return nil
	}

	return rd.(*order.Order)
}

func (o *BitfinexOrders) GetAll() []*order.Order {
	rs := make([]*order.Order, 0)

	o.orders.Range(func(key, value interface{}) bool {
		rs = append(rs, value.(*order.Order))
		return true
	})

	return rs
}

func BitfinexOrderToModel(o *order.Order) *models.Order {
	rs := &models.Order{
		ID:             fmt.Sprint(o.ID),
		Price:          o.Price,
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
	case "TRAILING STOP", "EXCHANGE TRAILING STOP":
		rs.Type = models.OrderTypeTrailingStop
	default:
		rs.Type = models.OrderTypeUnknown
	}

	rs.Meta["Exchange"] = ExchangeTypeBitfinex
	rs.Meta["Status"] = o.Status
	rs.Meta["Type"] = o.Type

	return rs
}

type BitfinexPositions struct {
	positions sync.Map
}

func (o *BitfinexPositions) Add(ps *position.Position) {
	o.positions.Store(ps.Id, ps)
}

func (o *BitfinexPositions) Get(positionID int64) *position.Position {
	wallet, ok := o.positions.Load(positionID)

	if !ok {
		return nil
	}

	return wallet.(*position.Position)
}

func (o *BitfinexPositions) Delete(positionID int64) *position.Position {
	wallet, ok := o.positions.LoadAndDelete(positionID)

	if !ok {
		return nil
	}

	return wallet.(*position.Position)
}

func (o *BitfinexPositions) GetAll() []*position.Position {
	rs := make([]*position.Position, 0)

	o.positions.Range(func(key, value interface{}) bool {
		rs = append(rs, value.(*position.Position))
		return true
	})

	return rs
}

func BitfinexPositionToModel(p *position.Position) *models.Position {
	rs := &models.Position{
		ID:                   fmt.Sprint(p.Id),
		Pair:                 p.Symbol,
		Price:                p.BasePrice,
		Amount:               p.Amount,
		LiqPrice:             p.LiquidationPrice,
		MarginLevel:          p.Leverage,
		ProfitLoss:           p.ProfitLoss,
		ProfitLossPercentage: p.ProfitLossPercentage,
		Meta:                 make(map[string]interface{}),
	}

	rs.Meta["Exchange"] = ExchangeTypeBitfinex

	return rs
}
