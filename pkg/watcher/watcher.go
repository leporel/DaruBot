package watcher

import (
	"fmt"
	"reflect"
	"sync"
)

type event struct {
	Type    EventType
	Payload interface{}
}

type EventType uint8
type EventsMap map[EventType]reflect.Type

func NewEvent(eType EventType, payload interface{}) *event {
	return &event{
		Type:    eType,
		Payload: payload,
	}
}

type Watcher struct {
	eventPipe      chan *event
	subscribeTypes []EventType
}

func newWatcher(eTypes ...EventType) *Watcher {
	return &Watcher{
		eventPipe:      make(chan *event, 10),
		subscribeTypes: eTypes,
	}
}

func (w *Watcher) Listen() <-chan *event {
	return w.eventPipe
}

func (w *Watcher) isListenType(eType EventType) bool {
	if len(w.subscribeTypes) == 0 {
		return true // Pass all event
	}

	for _, t := range w.subscribeTypes {
		if t == eType {
			return true
		}
	}

	return false
}

type Manager struct {
	mu       *sync.Mutex
	watchers map[string]*Watcher
	events   EventsMap
}

func NewWatcherManager() *Manager {
	return &Manager{
		mu:       &sync.Mutex{},
		watchers: make(map[string]*Watcher),
		events:   nil,
	}
}

func (w *Manager) RegisterEvents(eventsList map[EventType]reflect.Type) {
	w.events = eventsList
}

func (w *Manager) SupportEvents() EventsMap {
	return w.events
}

func (w *Manager) New(name string, eTypes ...EventType) (*Watcher, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.watchers[name]; ok {
		return nil, fmt.Errorf("watcher name '%s' already exist", name)
	}

	wh := newWatcher(eTypes...)

	w.watchers[name] = wh

	return wh, nil
}

func (w *Manager) Emmit(evt *event) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.checkType(evt); err != nil {
		return err
	}

	for _, wh := range w.watchers {
		if wh.isListenType(evt.Type) {
			wh.eventPipe <- evt
		}
	}

	return nil
}

func (w *Manager) checkType(evt *event) error {
	if t, ok := w.events[evt.Type]; ok {
		et := reflect.TypeOf(evt.Payload)
		if t != et {
			return fmt.Errorf("event contain wrong payload data: got (%s), expected (%s)\n", et, t)
		}
	}
	return nil
}

func (w *Manager) Remove(name string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	wh, ok := w.watchers[name]
	if !ok {
		return false
	}

	close(wh.eventPipe)

	delete(w.watchers, name)

	return true
}
