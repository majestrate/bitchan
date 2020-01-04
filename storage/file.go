package storage

type fileStorage struct {
	RootDir string
}

func (f *fileStorage) SetRoot(path string) {
	f.RootDir = path
}

func (f *fileStorage) GetRoot() string {
	return f.RootDir
}

func newFileStore() *fileStorage {
	return &fileStorage{}
}
