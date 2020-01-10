package db

import (
	"github.com/majestrate/bitchan/model"
)

type Facade interface {
	MigrateAll() error
	GetThreads(limit int) ([]model.ThreadInfo, error)
}
