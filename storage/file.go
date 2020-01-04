package storage

import (
	"os"
)

type fileStorage struct {
	RootDir string
}

func (f *fileStorage) SetRoot(path string) {
	f.RootDir = path
	os.Mkdir(f.RootDir, os.FileMode(0700))
}

func (f *fileStorage) GetRoot() string {
	return f.RootDir
}

func newFileStore() *fileStorage {
	return &fileStorage{}
}
