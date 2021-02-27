package storage

import (
	"DaruBot/internal/config"
	"DaruBot/pkg/errors"
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/json"
)

type localStorage struct {
	db *storm.DB
}

func New(cfg config.Configurations) (*localStorage, error) {
	if cfg.Storage.Local.Path == "" {
		return nil, errors.WrapMessage(ErrBadStoragePath, fmt.Sprintf("path: %s", cfg.Storage.Local.Path))
	}

	db, err := storm.Open(cfg.Storage.Local.Path, storm.Codec(json.Codec))
	if err != nil {
		return nil, err
	}

	return &localStorage{
		db: db,
	}, nil
}

func (s *localStorage) Stop() error {
	return s.db.Close()
}

func (s *localStorage) saveKV(bucket, key string, data interface{}) error {
	return s.db.Set(bucket, key, data)
}

func (s *localStorage) loadKV(bucket, key string, data interface{}) error {
	err := s.db.Get(bucket, key, data)
	return err
}

/*
	INFO Maybe if used storm TypeStore, do db.ReIndex(&Data{}) always when invoke ProvideStorage()
*/

func (s *localStorage) ProvideStatsStorage() (StatsStorage, error) {
	ss := &statsStore{
		s:      s,
		bucket: "stats",
	}

	return ss, nil
}

func (s *localStorage) ProvideCustomStorage(bucket string) (CustomStorage, error) {
	cs := &customStore{
		s:      s,
		bucket: bucket,
	}

	return cs, nil
}
