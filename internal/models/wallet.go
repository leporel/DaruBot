package models

import (
	"sync"
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
}

func (w *Wallets) Add(wallet *WalletCurrency) {
	w.wallets.Store(wallet.Name, wallet)
}

func (w *Wallets) Get(currency string) *WalletCurrency {
	wallet, ok := w.wallets.Load(currency)

	if !ok {
		return nil
	}

	return wallet.(*WalletCurrency)
}

func (w *Wallets) Delete(currency string) *WalletCurrency {
	wallet, ok := w.wallets.LoadAndDelete(currency)

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
