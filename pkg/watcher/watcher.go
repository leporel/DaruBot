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

type ModuleType string

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

// NewEventType create new event type.
// If dataType is nil, then type of event will be not compare to type of payload.
// May be helpful if  you are not restrict type of payload or you dont wanna use reflect.
func NewEventType(moduleType ModuleType, name string, dataType interface{}) *eventHead {
	if name == "" {
		panic("empty name")
	}

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
	eventPipe       chan *event // lazy init
	subscribeEvents []EventHead
	handler         func(*event)
	emitter         string
	module          ModuleType
}

func newWatcher(mType ModuleType, emitterName string, heads ...EventHead) *Watcher {
	return &Watcher{
		subscribeEvents: heads,
		emitter:         emitterName,
		module:          mType,
	}
}

// Listen make and return of event channel
// If channel not have readers, then will be runtime deadlock when buffer is full
func (w *Watcher) Listen() <-chan *event {
	if w.eventPipe == nil {
		w.eventPipe = make(chan *event, 30)
	}
	return w.eventPipe
}

// SetHandler set func to handle events from this watcher
func (w *Watcher) SetHandler(handler func(*event)) error {
	if handler == nil {
		return fmt.Errorf("handler is nil")
	}
	w.handler = handler
	return nil
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

// RegisterEvents register events list type for module, it will be used only for output from SupportEvents
func (w *Manager) RegisterEvents(moduleName string, events EventsMap) error {
	if _, ok := w.modulesTypes[moduleName]; ok {
		return fmt.Errorf("this module name already exist")
	}
	w.modulesTypes[moduleName] = events
	return nil
}

// SupportEvents returns the list of events which may be send by this module
func (w *Manager) SupportEvents(moduleName string) EventsMap {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.modulesTypes[moduleName]
}

// New return new watcher
// listenModule, listenModuleName, listenEvents can be empty
func (w *Manager) New(watcherName string, listenModule ModuleType, listenModuleName string, listenEvents ...EventHead) (*Watcher, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.watchers[watcherName]; ok {
		return nil, fmt.Errorf("watcher name '%s' already exist", watcherName)
	}

	wh := newWatcher(listenModule, listenModuleName, listenEvents...)

	w.watchers[watcherName] = wh

	return wh, nil
}

// MustNew return new watcher or panic
// listenModule, listenModuleName, listenEvents can be empty
func (w *Manager) MustNew(watcherName string, listenModule ModuleType, listenModuleName string, listenEvents ...EventHead) *Watcher {
	wh, err := w.New(watcherName, listenModule, listenModuleName, listenEvents...)
	if err != nil {
		panic(fmt.Errorf("watcher name '%s' already exist", watcherName))
	}

	return wh
}

// Emmit send event to watchers
func (w *Manager) Emmit(evt *event) error {
	if err := w.checkType(evt); err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, wh := range w.watchers {
		if wh.module != "" && wh.module != evt.GetModuleType() {
			continue
		}
		if wh.emitter != "" && wh.emitter != evt.ModuleName {
			continue
		}
		if wh.isListenType(evt.EventHead) {
			if wh.eventPipe != nil {
				wh.eventPipe <- evt
			}
			if wh.handler != nil {
				wh.handler(evt)
			}
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

// Remove watcher
func (w *Manager) Remove(watcherName string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	wh, ok := w.watchers[watcherName]
	if !ok {
		return false
	}
	if wh.eventPipe != nil {
		close(wh.eventPipe)
	}

	delete(w.watchers, watcherName)

	return true
}
