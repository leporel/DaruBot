package watcher

import (
	"fmt"
	"reflect"
	"sync"
)

type event struct {
	EventHead
	ModuleName string
	Payload    interface{}
}

func (e *event) Is(c EventHead) bool {
	return e.EventHead == c
}

type ModuleType uint8

type EventHead interface {
	GetModuleType() ModuleType
	GetEventName() string
}

type eventHead struct {
	ModuleType  ModuleType
	EventName   string
	payloadType reflect.Type
}

func (e *eventHead) GetModuleType() ModuleType {
	return e.ModuleType
}

func (e *eventHead) GetEventName() string {
	return e.EventName
}

func NewEvent(moduleType ModuleType, name string, dataType interface{}) *eventHead {
	var pT reflect.Type
	if dataType != nil {
		dt := reflect.TypeOf(dataType)
		pT = dt
		if dt.Kind() == reflect.Ptr {
			pT = dt.Elem()
		}
	}

	return &eventHead{
		ModuleType:  moduleType,
		EventName:   name,
		payloadType: pT,
	}
}

type EventsMap []EventHead

func BuildEvent(head EventHead, moduleName string, payload interface{}) *event {
	return &event{
		EventHead:  head,
		ModuleName: moduleName,
		Payload:    payload,
	}
}

type Watcher struct {
	eventPipe       chan *event
	subscribeEvents []EventHead
}

func newWatcher(heads ...EventHead) *Watcher {
	return &Watcher{
		eventPipe:       make(chan *event, 10),
		subscribeEvents: heads,
	}
}

func (w *Watcher) Listen() <-chan *event {
	return w.eventPipe
}

func (w *Watcher) isListenType(head EventHead) bool {
	if len(w.subscribeEvents) == 0 {
		return true // Pass all event
	}

	for _, t := range w.subscribeEvents {
		if t == head {
			return true
		}
	}

	return false
}

type Manager struct {
	mu           *sync.Mutex
	watchers     map[string]*Watcher
	modulesTypes map[string]EventsMap
}

func NewWatcherManager() *Manager {
	return &Manager{
		mu:           &sync.Mutex{},
		watchers:     make(map[string]*Watcher),
		modulesTypes: make(map[string]EventsMap),
	}
}

func (w *Manager) RegisterEvents(moduleName string, events EventsMap) {
	w.modulesTypes[moduleName] = events
}

func (w *Manager) SupportEvents(moduleName string) EventsMap {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.modulesTypes[moduleName]
}

func (w *Manager) New(name string, heads ...EventHead) (*Watcher, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.watchers[name]; ok {
		return nil, fmt.Errorf("watcher name '%s' already exist", name)
	}

	wh := newWatcher(heads...)

	w.watchers[name] = wh

	return wh, nil
}

func (w *Manager) MustNew(name string, heads ...EventHead) *Watcher {
	wh, err := w.New(name, heads...)
	if err != nil {
		panic(fmt.Errorf("watcher name '%s' already exist", name))
	}

	return wh
}

func (w *Manager) Emmit(evt *event) error {
	if err := w.checkType(evt); err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, wh := range w.watchers {
		if wh.isListenType(evt.EventHead) {
			wh.eventPipe <- evt
		}
	}

	return nil
}

func (w *Manager) checkType(evt *event) error {
	regT := evt.EventHead.(*eventHead).payloadType
	plT := reflect.TypeOf(evt.Payload)
	if regT != nil && regT != plT {
		return fmt.Errorf("event contain wrong payload type: got (%s), expected (%s)\n", plT, regT)
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
