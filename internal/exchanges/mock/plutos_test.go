package mock

import "DaruBot/internal/models"

func newPlutos(w *models.Wallets, currency string) *Plutos {
	ors := make([]models.Order, 0)
	pos := make([]models.Position, 0)

	if w == nil {
		w = &models.Wallets{}
		w.WalletType = models.WalletTypeNone
		wCur := &models.WalletCurrency{
			Name:       currency,
			WalletType: models.WalletTypeNone,
			Balance:    1000,
		}
		if len(pos) == 0 && len(ors) == 0 {
			wCur.Available = wCur.Balance
		}
		w.Update(wCur)
	}

	return NewPlutos(5, 0.2, currency, w, ors, pos)
}
