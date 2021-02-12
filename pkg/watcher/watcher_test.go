package watcher

import (
	"DaruBot/internal/models"
	"reflect"
	"testing"
)

func BenchmarkType(b *testing.B) {
	tp := reflect.TypeOf(&models.WalletCurrency{}).Elem()
	str := tp.String()
	for i := 0; i < b.N; i++ {
		tp2 := reflect.TypeOf(&models.WalletCurrency{}).Elem()
		if str == tp2.String() {
			// cool
		}
	}
}

func BenchmarkTypeString(b *testing.B) {
	tp := reflect.TypeOf(&models.WalletCurrency{}).Elem()
	for i := 0; i < b.N; i++ {
		tp2 := reflect.TypeOf(&models.WalletCurrency{}).Elem()
		if tp == tp2 {
			// cool
		}
	}
}

//BenchmarkType-16          	16899075	        73.1 ns/op
//BenchmarkTypeString-16    	20309344	        60.0 ns/op
