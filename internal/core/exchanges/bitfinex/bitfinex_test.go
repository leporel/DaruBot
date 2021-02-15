package bitfinex

import (
	"DaruBot/internal/config"
	"DaruBot/internal/core/exchanges"
	"DaruBot/internal/models"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/watcher"
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func Test_checkPair(t *testing.T) {
	type args struct {
		pair   string
		margin bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "exchange",
			args: args{
				pair:   "tBTCUSD",
				margin: false,
			},
			wantErr: false,
		},
		{
			name: "margin",
			args: args{
				pair:   "tBTCUSD",
				margin: true,
			},
			wantErr: false,
		},
		{
			name: "margin",
			args: args{
				pair:   "BTCUSD",
				margin: true,
			},
			wantErr: true,
		},
		{
			name: "margin",
			args: args{
				pair:   "tBTCBTC",
				margin: true,
			},
			wantErr: true,
		},
		{
			name: "margin",
			args: args{
				pair:   "tTESTBTC:TESTUSD",
				margin: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkPair(tt.args.pair, tt.args.margin); (err != nil) != tt.wantErr {
				t.Errorf("checkPair() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newBf(level logger.Level) (*Bitfinex, func(), error) {
	lg := logger.New(os.Stdout, level)
	ctx, finish := context.WithCancel(context.Background())

	wManager := watcher.NewWatcherManager()

	bf, err := newBitfinex(ctx, config.Config, wManager, lg)

	return bf, finish, err
}

func startWatcher(t *testing.T, bf *Bitfinex) func() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		wh := bf.watchers.MustNew("all_events")

		for {
			select {
			case evt := <-wh.Listen():
				if evt.Is(models.EventError) {
					t.Fatalf("error: %v", evt.Payload)
				}
				//t.Logf("event type: %v(%v), payload: [%#v] \n", EventToString(evt.Head), evt.Head, evt.Payload)
				fmt.Printf("event: %v(%v), payload: [%#v] \n", evt.GetModuleType(), evt.GetEventName(), evt.Payload)
			case <-ctx.Done():
				bf.watchers.Remove("all_events")
				return
			}
		}
	}()

	return cancel
}

func Test_BitfinexConnect(t *testing.T) {
	bf, finish, err := newBf(logger.DebugLevel)
	if err != nil {
		t.Errorf("error = %v", err)
	}

	watcherEnd := startWatcher(t, bf)
	defer watcherEnd()

	err = bf.Connect()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	time.Sleep(2 * time.Second)
	finish()
	time.Sleep(1 * time.Second)
}

func Test_BitfinexOrderAndPositionsList(t *testing.T) {
	bf, finish, err := newBf(logger.DebugLevel)
	if err != nil {
		t.Errorf("error = %v", err)
	}

	err = bf.Connect()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	time.Sleep(2 * time.Second)

	rds, err := bf.GetOrders()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	pns, err := bf.GetPositions()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	for _, rd := range rds {
		t.Logf("Order: %#v \n", rd)
	}
	for _, ps := range pns {
		t.Logf("Position: %#v \n", ps)
	}

	finish()
	time.Sleep(1 * time.Second)
}

func Test_BitfinexOrder(t *testing.T) {
	bf, finish, err := newBf(logger.DebugLevel)
	if err != nil {
		t.Errorf("error = %v", err)
	}

	watcherEnd := startWatcher(t, bf)
	defer watcherEnd()

	err = bf.Connect()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	time.Sleep(2 * time.Second)

	newOrder := &models.PutOrder{
		InternalID: fmt.Sprint(time.Now().Unix() / 1000),
		Pair:       "tTESTBTC:TESTUSD",
		Type:       models.OrderTypeLimit,
		Amount:     0.001,
		Price:      300,
		StopPrice:  0,
		Margin:     false,
	}

	order, err := bf.PutOrder(newOrder)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	t.Logf("Created order: %#v", order)

	time.Sleep(2 * time.Second)

	order, err = bf.UpdateOrder(order.ID, 500, 0, 0.002)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	t.Logf("Updated order: %#v", order)

	time.Sleep(2 * time.Second)

	err = bf.CancelOrder(order)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	t.Logf("Canceled order: %#v", order)

	time.Sleep(3 * time.Second)
	finish()
	time.Sleep(1 * time.Second)
}

func Test_BitfinexOrderCancelError(t *testing.T) {
	bf, finish, err := newBf(logger.DebugLevel)
	if err != nil {
		t.Errorf("error = %v", err)
	}

	watcherEnd := startWatcher(t, bf)
	defer watcherEnd()

	err = bf.Connect()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	time.Sleep(2 * time.Second)

	newOrder := &models.Order{
		ID:         "34534534",
		InternalID: fmt.Sprint(time.Now().Unix() / 1000),
	}

	err = bf.CancelOrder(newOrder)
	if err == nil {
		t.Fatalf("dont recieved error")
	}
	if err == exchanges.ErrResultTimeOut {
		t.Fatalf("request timeout")
	}

	time.Sleep(1 * time.Second)
	finish()
	time.Sleep(1 * time.Second)
}

func Test_BitfinexTestPosition(t *testing.T) {
	bf, finish, err := newBf(logger.DebugLevel)
	if err != nil {
		t.Errorf("error = %v", err)
	}

	watcherEnd := startWatcher(t, bf)
	defer watcherEnd()

	err = bf.Connect()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	time.Sleep(2 * time.Second)

	newOrder := &models.PutOrder{
		InternalID: fmt.Sprint(time.Now().Unix() / 1000),
		Pair:       "tTESTBTC:TESTUSD",
		Type:       models.OrderTypeMarket,
		Amount:     0.001,
		Price:      0,
		StopPrice:  0,
		Margin:     true,
	}

	wh := bf.watchers.MustNew("new_position", models.EventPositionNew, models.EventPositionUpdate)
	defer bf.watchers.Remove("new_position")

	Timout := time.NewTimer(10 * time.Second)
	defer Timout.Stop()

	go func() {
		for {
			select {
			case pos := <-wh.Listen():
				if pos.Payload.(models.Position).Pair == newOrder.Pair {
					t.Logf("New position: %#v", pos)
					return
				}
			case <-Timout.C:
				t.Fatalf("wait new position time out")
			}
		}
	}()

	order, err := bf.PutOrder(newOrder)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	t.Logf("Created order: %#v", order)

	time.Sleep(1 * time.Second)

	ps, err := bf.GetPositions()
	if err != nil {
		t.Fatalf("error = %v", err)
	}

	if len(ps) == 0 {
		t.Fatalf("positions are not appeared %s", err)
	}

	pos, err := bf.ClosePosition(ps[0])
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	t.Logf("Position closed: %#v", pos)

	time.Sleep(3 * time.Second)
	finish()
	time.Sleep(1 * time.Second)
}

func Test_BitfinexTicker(t *testing.T) {
	bf, finish, err := newBf(logger.DebugLevel)
	if err != nil {
		t.Errorf("error = %v", err)
	}

	watcherEnd := startWatcher(t, bf)
	defer watcherEnd()

	err = bf.Connect()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	ticker, err := bf.GetTicker("tTESTBTC:TESTUSD")
	if err != nil {
		t.Errorf("error = %v", err)
	}
	t.Logf("Ticker: %#v", ticker)

	t.Log("Subscribing")
	s, err := bf.SubscribeTicker("tTESTBTC:TESTUSD")
	if err != nil {
		t.Errorf("error = %v", err)
	}

	time.Sleep(15 * time.Second)

	err = bf.Unsubscribe(s)
	if err != nil {
		t.Errorf("error = %v", err)
	}

	time.Sleep(1 * time.Second)
	finish()
	time.Sleep(1 * time.Second)
}

func Test_BitfinexGetCandles(t *testing.T) {
	bf, finish, err := newBf(logger.DebugLevel)
	if err != nil {
		t.Errorf("error = %v", err)
	}

	watcherEnd := startWatcher(t, bf)
	defer watcherEnd()

	err = bf.Connect()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	candels, err := bf.GetCandles("tTESTBTC:TESTUSD", models.OneHour, time.Now().Add(-time.Hour*24*1), time.Now())
	if err != nil {
		t.Errorf("error = %v", err)
	}
	t.Logf("Candels: %#v len(%v)", candels, len(candels.Candles))
	for _, c := range candels.Candles {
		fmt.Printf("[%s] Open:%v Close:%v High:%v Low:%v Volume:%v \n",
			c.Date.Format("2006-01-02T15:04:05"),
			c.Open, c.Close, c.High, c.Low, c.Volume)
	}

	time.Sleep(1 * time.Second)
	finish()
	time.Sleep(1 * time.Second)
}
