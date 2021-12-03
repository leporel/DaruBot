package bitfinex

import (
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/order"
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
