package exchanges

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/tools"
	"fmt"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/order"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/position"
	"sync"
	"time"
)

type BitfinexOrders struct {
	orders     sync.Map
	lastUpdate time.Time
}

func (o *BitfinexOrders) Add(rd *order.Order) {
	o.orders.Store(rd.ID, rd)
	o.lastUpdate = time.Now()
}

func (o *BitfinexOrders) Get(orderID int64) *order.Order {
	rd, ok := o.orders.Load(orderID)

	if !ok {
		return nil
	}

	return rd.(*order.Order)
}

func (o *BitfinexOrders) Delete(orderID int64) *order.Order {
	o.lastUpdate = time.Now()
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

func (o *BitfinexOrders) Clear() {
	o.orders = sync.Map{}
	o.lastUpdate = time.Now()
}

func (o *BitfinexOrders) LastUpdate() time.Time {
	return o.lastUpdate
}

func BitfinexOrderToModel(or interface{}) (*models.Order, bool) {
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

	rs.Meta["Exchange"] = ExchangeTypeBitfinex
	rs.Meta["Status"] = o.Status
	rs.Meta["Type"] = o.Type
	rs.Meta["Flags"] = o.Flags

	return rs, true
}

type BitfinexPositions struct {
	positions  sync.Map
	lastUpdate time.Time
}

func (o *BitfinexPositions) Add(ps *position.Position) {
	o.positions.Store(ps.Id, ps)
	o.lastUpdate = time.Now()
}

func (o *BitfinexPositions) Get(positionID int64) *position.Position {
	wallet, ok := o.positions.Load(positionID)

	if !ok {
		return nil
	}

	return wallet.(*position.Position)
}

func (o *BitfinexPositions) Delete(positionID int64) *position.Position {
	o.lastUpdate = time.Now()
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

func (o *BitfinexPositions) Clear() {
	o.positions = sync.Map{}
	o.lastUpdate = time.Now()
}

func (o *BitfinexPositions) LastUpdate() time.Time {
	return o.lastUpdate
}

func BitfinexPositionToModel(po interface{}) (*models.Position, bool) {
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
	rs.Meta["Status"] = p.Status

	return rs, true
}
