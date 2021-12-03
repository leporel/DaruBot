package mock

import (
	"fmt"
	"sync"
	"time"
)

type TheWorld struct {
	lock *sync.Mutex

	timePass time.Duration
	from     time.Time
	to       time.Time
	tick     time.Duration

	ch   chan time.Time
	wait chan interface{}

	work bool
}

func NewTheWorld(from, to time.Time, tick time.Duration) *TheWorld {
	max := to.Sub(from)
	lock := &sync.Mutex{}

	stand := &TheWorld{
		lock:     lock,
		timePass: 0,
		from:     from,
		to:       to,
		tick:     tick,
		ch:       nil,
		wait:     make(chan interface{}, 1),
		work:     false,
	}

	lock.Lock()

	go func() {
		<-stand.wait
		fmt.Println("MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA MUDA!!!!")
		for {
			if stand.work == false {
				return
			}
			lock.Lock()
			time.Sleep(stand.tick)
			if stand.timePass > max {
				return
			}
			stand.timePass = stand.timePass + time.Minute
			if stand.ch != nil {
				stand.ch <- stand.CurrentTime()
			}
			lock.Unlock()
		}
	}()

	return stand
}

func (w *TheWorld) Run() {
	if w.wait != nil {
		close(w.wait)
		w.wait = nil
		w.work = true
	}
	w.TimeStart()
}

func (w *TheWorld) Stop() {
	w.work = false
}

func (w *TheWorld) SetTick(tick time.Duration) {
	w.tick = tick
}

// TimeStart Zero
func (w *TheWorld) TimeStart() {
	if w.wait != nil {
		return
	}
	w.lock.Unlock()
}

// TimeStop ZA WARUDO!!!!
func (w *TheWorld) TimeStop() {
	if w.wait != nil {
		return
	}
	w.lock.Lock()
}

func (w *TheWorld) GetChan() <-chan time.Time {
	if w.ch == nil {
		w.ch = make(chan time.Time, 1)
	}
	return w.ch
}

func (w *TheWorld) CurrentTime() time.Time {
	return w.from.Add(w.timePass)
}
