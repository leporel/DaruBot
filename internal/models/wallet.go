package models

import (
	"sync"
	"time"
)

type BalanceUSD struct {
	Total    float64
	NetWorth float64
}

type WalletType uint8

const (
	WalletTypeNone WalletType = iota
	WalletTypeExchange
	WalletTypeMargin
)

type WalletCurrency struct {
	Name       string
	WalletType WalletType
	Balance    float64
	Available  float64
}

type Wallets struct {
	WalletType WalletType
	wallets    sync.Map
	lastUpdate time.Time
}

func (w *Wallets) Add(wallet *WalletCurrency) {
	w.wallets.Store(wallet.Name, wallet)
	w.lastUpdate = time.Now()
}

func (w *Wallets) Get(symbol string) *WalletCurrency {
	wallet, ok := w.wallets.Load(symbol)

	if !ok {
		return nil
	}

	return wallet.(*WalletCurrency)
}

func (w *Wallets) Delete(symbol string) *WalletCurrency {
	w.lastUpdate = time.Now()
	wallet, ok := w.wallets.LoadAndDelete(symbol)

	if !ok {
		return nil
	}

	return wallet.(*WalletCurrency)
}

func (w *Wallets) GetAll() []*WalletCurrency {
	rs := make([]*WalletCurrency, 0)

	w.wallets.Range(func(key, value interface{}) bool {
		rs = append(rs, value.(*WalletCurrency))
		return true
	})

	return rs
}

func (w *Wallets) Clear() {
	w.wallets = sync.Map{}
	w.lastUpdate = time.Now()
}

func (w *Wallets) LastUpdate() time.Time {
	return w.lastUpdate
}
