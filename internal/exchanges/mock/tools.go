package mock

import (
	"DaruBot/internal/models"
	"fmt"
	"github.com/markcheno/go-quote"
	"sync"
	"time"
)

var (
	timeFormat  = "2006-01-02 15:04"
	timeFormatD = "2006-01-02"
)

type TheWorld struct {
	lock *sync.Mutex

	timePass time.Duration
	from     time.Time
	to       time.Time
	tick     time.Duration

	ch   chan time.Time
	wait chan interface{}
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
	}

	lock.Lock()

	go func() {
		<-stand.wait
		for {
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
	close(w.wait)
	w.wait = nil
	w.TimeStart()
}

func (w *TheWorld) SetTick(tick time.Duration) {
	w.tick = tick
}

// TimeStop Zero
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

func getCandle(q *quote.Quote, t time.Time) *models.Candle {
	if len(q.Date) == 0 {
		return nil
	}

	if t.IsZero() {
		return models.QuoteToModel(q, q.Symbol, len(q.Date)-1, "")
	}

	if len(q.Date) < 2 {
		return nil
	}

	d := q.Date[1].Sub(q.Date[0])

	for i := 0; i < len(q.Date); i++ {
		if q.Date[i].Sub(t) <= d {
			return models.QuoteToModel(q, q.Symbol, i, "")
		}
	}

	return nil
}

func quoteFormat(t time.Time, format string) string {
	return t.Format(format)
}

func getDailyKey(from, to time.Time, symbol string) string {
	return fmt.Sprintf("%s_%s/%s_%s", models.OneDay, from.Format("2006-01-02"), to.Format("2006-01-02"), symbol)
}

func getMinuteKey(currentTime time.Time, symbol string) string {
	return fmt.Sprintf("%s_%s_%s", models.OneMinute, currentTime.Format("2006-01-02"), symbol)
}
