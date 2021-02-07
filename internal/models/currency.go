package models

type Currency struct {
	Pair  string
	Price float64
	State *CurrencyState
}

// CurrencyState https://docs.bitfinex.com/reference#rest-public-tickers
type CurrencyState struct {
	High    float64
	Low     float64
	Volume  float64
	BID     float64
	BIDSize float64
	ASK     float64
	ASKSize float64
}

func (c *Currency) GetHigh() (float64, bool) {
	if c.State == nil || c.State.High == 0 {
		return 0, false
	}

	return c.State.High, true
}

func (c *Currency) GetLow() (float64, bool) {
	if c.State == nil || c.State.Low == 0 {
		return 0, false
	}

	return c.State.Low, true
}

func (c *Currency) GetVolume() (float64, bool) {
	if c.State == nil || c.State.Volume == 0 {
		return 0, false
	}

	return c.State.Volume, true
}

func (c *Currency) GetBID() (float64, bool) {
	if c.State == nil || c.State.BID == 0 {
		return 0, false
	}

	return c.State.BID, true
}

func (c *Currency) GetBIDSize() (float64, bool) {
	if c.State == nil || c.State.BIDSize == 0 {
		return 0, false
	}

	return c.State.BIDSize, true
}

func (c *Currency) GetASK() (float64, bool) {
	if c.State == nil || c.State.ASK == 0 {
		return 0, false
	}

	return c.State.ASK, true
}

func (c *Currency) GetASKSize() (float64, bool) {
	if c.State == nil || c.State.ASKSize == 0 {
		return 0, false
	}

	return c.State.ASKSize, true
}
