package mock

import (
	"DaruBot/internal/models"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type theWorld struct {
	lock *sync.Mutex

	timePass time.Duration
	from     time.Time
	to       time.Time

	ch   chan time.Time
	wait chan interface{}
}

func newTheWorld(from, to time.Time) *theWorld {
	max := to.Sub(from)
	lock := &sync.Mutex{}

	stand := &theWorld{
		lock:     lock,
		timePass: 0,
		from:     from,
		to:       to,
		ch:       nil,
		wait:     make(chan interface{}),
	}

	lock.Lock()

	go func() {
		<-stand.wait
		for {
			lock.Lock()
			time.Sleep(1 * time.Millisecond)
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

func (w *theWorld) Run() {
	close(w.wait)
	w.wait = nil
	w.TimeStart()
}

// TimeStop Zero
func (w *theWorld) TimeStart() {
	if w.wait != nil {
		return
	}
	w.lock.Unlock()
}

// TimeStop ZA WARUDO!!!!
func (w *theWorld) TimeStop() {
	if w.wait != nil {
		return
	}
	w.lock.Lock()
}

func (w *theWorld) GetChan() <-chan time.Time {
	if w.ch == nil {
		w.ch = make(chan time.Time, 1)
	}
	return w.ch
}

func (w *theWorld) CurrentTime() time.Time {
	return w.from.Add(w.timePass)
}

func getRandFloat(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())
	n := rand.Int63n(int64(max - min))
	r := min + float64(n)
	return r
}

var timeFormat = "2006-01-02 15:04"

func timeString(t time.Time) string {
	return t.Format(timeFormat)
}

func getDailyKey(from, to time.Time, pair string) string {
	return fmt.Sprintf("%s_%s/%s_%s", models.OneDay, from.Format("2006-01-02"), to.Format("2006-01-02"), pair)
}

func getMinuteKey(currentTime time.Time, pair string) string {
	return fmt.Sprintf("%s_%s_%s", models.OneMinute, currentTime.Format("2006-01-02"), pair)
}
