package storage

import (
	"DaruBot/internal/models"
	"DaruBot/storage/udt"
)

type statsStore struct {
	s      *localStorage
	bucket string
}

type StatsStorage interface {
	SaveStats(*models.Stats) error
	LoadStats() (*models.Stats, error)
}

func (s *statsStore) SaveStats(stats *models.Stats) error {
	data := &udt.Stats{
		TotalLoss:   stats.TotalLoss,
		TotalProfit: stats.TotalProfit,
		TotalTrades: stats.TotalTrades,
	}

	return s.s.saveKV(s.bucket, data.Version(), data)
}

func (s *statsStore) LoadStats() (*models.Stats, error) {
	data := &udt.Stats{}
	err := s.s.loadKV(s.bucket, data.Version(), data)
	if err != nil {
		return nil, err
	}

	return &models.Stats{
		TotalLoss:   data.TotalLoss,
		TotalProfit: data.TotalProfit,
		TotalTrades: data.TotalTrades,
	}, nil
}
