package mock

import (
	"github.com/markcheno/go-quote"
	"github.com/sanity-io/litter"
	"testing"
	"time"
)

func TestPairList(t *testing.T) {
	list, err := quote.NewMarketList(market)
	if err != nil {
		t.Fatal(err)
	}

	litter.Dump(list)
}

func TestDailyTicker(t *testing.T) {
	start := time.Now().UTC().Add(-time.Hour * 24)
	end := time.Now()
	q, err := quoteFrom(testPair, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"), quote.Daily)
	if err != nil {
		t.Fatal(err)
	}

	litter.Dump(q)
}

func TestMinuteTicker(t *testing.T) {
	start := time.Now().UTC().Add(-time.Minute * 2)
	end := time.Now().UTC()
	q, err := quoteFrom(testPair, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"), quote.Min1)
	if err != nil {
		t.Fatal(err)
	}

	litter.Dump(q)
}
