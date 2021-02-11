package exchanges

import (
	"DaruBot/internal/config"
	"DaruBot/internal/models"
	"DaruBot/pkg/logger"
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

func newBitfinex(level logger.Level) (*BitFinex, func(), error) {
	lg := logger.New(os.Stdout, level)
	ctx, finish := context.WithCancel(context.Background())

	bf, err := NewBitFinex(ctx, config.Config, lg)

	return bf, finish, err
}

func startWatcher(t *testing.T, bf *BitFinex) func() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		wh := bf.RegisterWatcher("all_events")

		for {
			select {
			case evt := <-wh.Listen():
				if evt.Type == EventError {
					t.Errorf("error: %v", evt.Payload)
				}
				t.Logf("event type: %v(%v), payload: [%#v] \n", EventToString(evt.Type), evt.Type, evt.Payload)
				//fmt.Printf("\nevent type: %v, payload: [%#v] \n\n", evt.Type, evt.Payload)
			case <-ctx.Done():
				bf.RemoveWatcher("all_events")
				return
			}
		}
	}()

	return cancel
}

func Test_BitfinexConnect(t *testing.T) {
	bf, finish, err := newBitfinex(logger.DebugLevel)
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
	bf, finish, err := newBitfinex(logger.DebugLevel)
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
	bf, finish, err := newBitfinex(logger.DebugLevel)
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

	time.Sleep(3 * time.Second)

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
	bf, finish, err := newBitfinex(logger.DebugLevel)
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
	if err == ErrResultTimeOut {
		t.Fatalf("request timeout")
	}

	time.Sleep(1 * time.Second)
	finish()
	time.Sleep(1 * time.Second)
}
