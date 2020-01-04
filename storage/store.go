package storage

type Store interface {
	SetRoot(d string)
	GetRoot() string
}

func NewStorage() Store {
	return newFileStore()
}
