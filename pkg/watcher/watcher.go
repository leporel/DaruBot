package watcher

import "sync"

type event struct {
	Type    EventType
	Payload interface{}
}

type EventType uint8

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

type WatcherManager struct {
	mu       *sync.Mutex
	watchers map[string]*Watcher
}

func NewWatcherManager() *WatcherManager {
	return &WatcherManager{
		mu:       &sync.Mutex{},
		watchers: make(map[string]*Watcher),
	}
}

func (w *WatcherManager) New(name string, eTypes ...EventType) *Watcher {
	w.mu.Lock()
	defer w.mu.Unlock()

	wh := newWatcher(eTypes...)

	w.watchers[name] = wh

	return wh
}

func (w *WatcherManager) Emmit(evt *event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, wh := range w.watchers {
		if wh.isListenType(evt.Type) {
			wh.eventPipe <- evt
		}
	}
}

func (w *WatcherManager) Remove(name string) bool {
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
