package exchanges

import (
	"DaruBot/internal/config"
	"DaruBot/pkg/logger"
	"context"
	"github.com/sirupsen/logrus"
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

func newBitfinex() (*BitFinex, func(), error) {
	lg := logger.New(os.Stdout, logrus.DebugLevel)
	ctx, finish := context.WithCancel(context.Background())

	bf, err := NewBitFinex(ctx, config.Config, lg)

	return bf, finish, err
}

func Test_BitfinexConnect(t *testing.T) {
	bf, finish, err := newBitfinex()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	err = bf.Connect()
	if err != nil {
		t.Errorf("error = %v", err)
	}

	time.Sleep(2 * time.Second)
	finish()
	time.Sleep(1 * time.Second)
}

func Test_BitfinexOrderAndPositionsList(t *testing.T) {
	bf, finish, err := newBitfinex()
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
