package storage

type customStore struct {
	s      *localStorage
	bucket string
}

type CustomStorage interface {
	Save(key string, data interface{}) error
	Load(key string, to interface{}) error
}

func (c *customStore) Save(key string, data interface{}) error {
	return c.s.saveKV(c.bucket, key, data)
}

func (c *customStore) Load(key string, to interface{}) error {

	err := c.s.loadKV(c.bucket, key, to)
	if err != nil {
		return err
	}

	return nil
}
