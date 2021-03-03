package watcher

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type S struct {
	foo string
	bar string
}

func BenchmarkType(b *testing.B) {
	tp := reflect.TypeOf(&S{}).Elem()
	for i := 0; i < b.N; i++ {
		tp2 := reflect.TypeOf(&S{}).Elem()
		if tp == tp2 {
			// cool
		}
	}
}

func BenchmarkTypeString(b *testing.B) {
	tp := reflect.TypeOf(&S{}).Elem()
	str := tp.String()
	for i := 0; i < b.N; i++ {
		tp2 := reflect.TypeOf(&S{}).Elem()
		if str == tp2.String() {
			// cool
		}
	}
}

//BenchmarkType-16          	16899075	        73.1 ns/op
//BenchmarkTypeString-16    	20309344	        60.0 ns/op

func TestHandler(t *testing.T) {
	m := NewWatcherManager()

	eName := "tester"
	mName := ModuleType("module name")

	wh, err := m.New("test watcher", mName, eName)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Remove("test watcher")

	wait := make(chan bool, 1)
	handler := func(evt *event) {
		fmt.Printf("%#v\n", evt)
		wait <- true
	}

	err = wh.SetHandler(handler)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Emmit(&event{
		EventHead:  NewEventType(mName, "test event", nil),
		ModuleName: eName,
		Payload:    "meh",
	})
	if err != nil {
		t.Fatal(err)
	}

	tk := time.Tick(1 * time.Second)

	for {
		select {
		case <-wait:
			return
		case <-tk:
			t.Fatal("event not handled")
		}
	}
}
