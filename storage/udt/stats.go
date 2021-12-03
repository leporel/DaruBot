package udt

type Stats struct {
	TotalLoss   float64
	TotalProfit float64
	TotalTrades int
}

func (s *Stats) Version() string {
	return "v1"
}
