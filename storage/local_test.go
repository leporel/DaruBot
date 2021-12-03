package storage

import (
	"DaruBot/internal/config"
	"DaruBot/internal/models"
	"reflect"
	"testing"
)

func newLS() (*localStorage, error) {
	cfg := config.GetDefaultConfig()

	cfg.Storage.Local.Path = "../test_data/storage_test.db"
	return New(cfg)
}

func TestStats(t *testing.T) {

	s, err := newLS()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	ss, err := s.ProvideStatsStorage()
	if err != nil {
		t.Fatal(err)
	}

	stats := models.Stats{
		TotalLoss:   3.4,
		TotalProfit: 5.6,
		TotalTrades: 10,
	}

	err = ss.SaveStats(&stats)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := ss.LoadStats()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(&stats, loaded) {
		t.Fatalf("exepted %#v, got %#v", &stats, loaded)
	}
}

func TestCustomStore(t *testing.T) {
	s, err := newLS()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	cs, err := s.ProvideCustomStorage("custom1")
	if err != nil {
		t.Fatal(err)
	}

	type CustomData struct {
		Foo string
		Bar string
	}

	data := CustomData{
		Foo: "123",
		Bar: "456",
	}

	err = cs.Save("data", data)
	if err != nil {
		t.Fatal(err)
	}

	loaded := &CustomData{}

	err = cs.Load("data", loaded)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(&data, loaded) {
		t.Fatalf("exepted %#v, got %#v", &data, loaded)
	}
}
