package mock

import (
	"fmt"
	"sync"
	"time"
)

const (
	stepInterval = time.Minute
)

// Dio Implement time simulation
type Dio struct {
	lock  *sync.Mutex
	pause *sync.Mutex

	from, to, current time.Time
	tick              time.Duration

	ch   chan time.Time
	inProgress bool
	exit chan interface{}
}

// NewDio create new time emulation instance
func NewDio(from, to time.Time, tick time.Duration) *Dio {
	stand := &Dio{
		lock:     &sync.Mutex{},
		pause:    &sync.Mutex{},
		from:     from,
		to:       to,
		current:  from,
		tick:     tick,
		ch:       nil,
		inProgress:     false,
		exit:     make(chan interface{}, 1),
	}

	return stand
}

func (w *Dio) Run() {
	if !w.inProgress {
		go func() {
			w.inProgress = true
			fmt.Println("MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA!!!!")
			for {
				select {
				case <-w.exit:
					return
				default:
					w.pause.Lock() // FIXME change to waitGroup
					time.Sleep(w.tick)
					if w.current.After(w.to) {
						w.Done()
					}
					w.addStep(stepInterval)
	
					if w.ch != nil {
						w.ch <- w.Time()
					}
					w.pause.Unlock()
				}
			}
		}()
	}
	if w.current.After(w.to) {
		panic("Dio cant run him self again, create new Dio")
	}
}

func (w *Dio) addStep(interval time.Duration) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.current = w.current.Add(interval)
}

func (w *Dio) Done() {
	if !w.inProgress {
		return
	}
	w.exit <- struct{}{}
}

func (w *Dio) SetTick(tick time.Duration) {
	w.tick = tick
}

// TimeStart continue time traveling
// Zero...
func (w *Dio) Continue() {
	if !w.inProgress {
		return
	}
	w.pause.TryLock()
	w.pause.Unlock()
}

// TimeStop freeze time traveling
// ZA WARUDO!!!!
func (w *Dio) Pause() {
	if !w.inProgress {
		return
	}
	w.pause.Lock()
}

func (w *Dio) GetChan() <-chan time.Time {
	if w.ch == nil {
		w.ch = make(chan time.Time, 1)
	}
	return w.ch
}

// Time return current time
func (w *Dio) Time() time.Time {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.current
}
