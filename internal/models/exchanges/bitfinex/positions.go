package bitfinex

import (
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/position"
	"sync"
	"time"
)

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
