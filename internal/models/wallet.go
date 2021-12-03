package models

import (
	"sync"
	"time"
)

type BalanceUSD struct {
	Total    float64 // available usd
	NetWorth float64 // all stock and positions worth in usd and available usd
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

func (w *Wallets) Update(wallet *WalletCurrency) {
	w.wallets.Store(wallet.Name, wallet)
	w.lastUpdate = time.Now()
}

func (w *Wallets) Get(currencyName string) *WalletCurrency {
	wallet, ok := w.wallets.Load(currencyName)

	if !ok {
		return nil
	}

	return wallet.(*WalletCurrency)
}

func (w *Wallets) Delete(currencyName string) *WalletCurrency {
	w.lastUpdate = time.Now()
	wallet, ok := w.wallets.LoadAndDelete(currencyName)

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
