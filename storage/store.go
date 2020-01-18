package storage

type Store interface {
	SetRoot(d string)
	GetRoot() string
	GetAllTorrents() []string
}

func NewStorage() Store {
	return newFileStore()
}
