package storage

import (
	"os"
	"path/filepath"
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

func (f *fileStorage) GetAllTorrents() []string {
	matches, err := filepath.Glob(filepath.Join(f.RootDir, "*.torrent"))
	if err != nil {
		return []string{}
	}
	return matches
}

func newFileStore() *fileStorage {
	return &fileStorage{}
}
