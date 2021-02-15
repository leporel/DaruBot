package watcher

import (
	"reflect"
	"testing"
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
