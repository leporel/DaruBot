package models

import "sync"

type SubType uint8

const (
	SubTypeTicker SubType = iota
	SubTypeCandle
)

type Subscription struct {
	ID     string
	Symbol string
	Type   SubType
}

type Subscriptions struct {
	subscriptions sync.Map
}

func (s *Subscriptions) Add(sub *Subscription) {
	s.subscriptions.Store(sub.ID, sub)
}

func (s *Subscriptions) Get(ID string) *Subscription {
	sub, ok := s.subscriptions.Load(ID)

	if !ok {
		return nil
	}

	return sub.(*Subscription)
}

func (s *Subscriptions) Delete(ID string) *Subscription {
	sub, ok := s.subscriptions.LoadAndDelete(ID)

	if !ok {
		return nil
	}

	return sub.(*Subscription)
}

func (s *Subscriptions) GetAll() []*Subscription {
	rs := make([]*Subscription, 0)

	s.subscriptions.Range(func(key, value interface{}) bool {
		rs = append(rs, value.(*Subscription))
		return true
	})

	return rs
}

func (s *Subscriptions) Clear() {
	s.subscriptions = sync.Map{}
}
