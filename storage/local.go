package storage

type SimpleLocalStorage interface {
	Write(data []byte) error
	Read() ([]byte, error)
}

// TODO store summary

//func ProvideSimpleLocalStorage(c config.Configurations) SimpleLocalStorage {
//
//	return NewLocal(w)
//}
