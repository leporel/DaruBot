package bitfinex

import (
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

type BitfinexPositions struct {
	positions  sync.Map
	lastUpdate time.Time
}

func (o *BitfinexPositions) Add(ps *position.Position) {
	o.positions.Store(ps.Id, ps)
	o.lastUpdate = time.Now()
}

func (o *BitfinexPositions) Get(positionID int64) *position.Position {
	pos, ok := o.positions.Load(positionID)

	if !ok {
		return nil
	}

	return pos.(*position.Position)
}

func (o *BitfinexPositions) Delete(positionID int64) *position.Position {
	o.lastUpdate = time.Now()
	pos, ok := o.positions.LoadAndDelete(positionID)

	if !ok {
		return nil
	}

	return pos.(*position.Position)
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
