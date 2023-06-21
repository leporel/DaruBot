package mock

import (
	"DaruBot/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Ticker struct {
	Time   time.Time
	Symbol string
}

type Candle struct {
	Time   time.Time
	Symbol string
	Res    models.CandleResolution
}

type subscribeManager struct {
	subs    []*subscription
	seconds uint8
}

type subscription struct {
	id     string
	symbol string
	sType  models.SubType
	sRes   models.CandleResolution
}

func (p *subscribeManager) trigger(t time.Time, ch chan interface{}) {
	p.seconds++
	for _, s := range p.subs {
		switch s.sType {
		case models.SubTypeTicker:
			if p.seconds == 10 {
				ch <- &Ticker{
					Time:   t,
					Symbol: s.symbol,
				}
				p.seconds = 0
			}
		case models.SubTypeCandle:
			if checkResolutionInterval(s.sRes.ToDuration(), t) {
				ch <- &Candle{
					Time:   t,
					Symbol: s.symbol,
					Res:    s.sRes,
				}
			}
		}
	}
}

func (p *Plutos) SubscribeTicker(symbol string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	s := &subscription{
		id:     uuid.Must(uuid.NewUUID()).String(),
		symbol: symbol,
		sType:  models.SubTypeTicker,
		sRes:   "",
	}

	p.SubscribeManager.subs = append(p.SubscribeManager.subs, s)

	return s.id
}

func (p *Plutos) SubscribeCandle(symbol string, resolution models.CandleResolution) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	s := &subscription{
		id:     uuid.Must(uuid.NewUUID()).String(),
		symbol: symbol,
		sType:  models.SubTypeCandle,
		sRes:   resolution,
	}

	p.SubscribeManager.subs = append(p.SubscribeManager.subs, s)

	return s.id
}

func (p *Plutos) Unsubscribe(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, s := range p.SubscribeManager.subs {
		if s.id == id {
			p.SubscribeManager.subs = append(p.SubscribeManager.subs[:i], p.SubscribeManager.subs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("not found subscription")
}
