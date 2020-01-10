package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/majestrate/bitchan/model"
)

type pqFacade struct {
	db *sql.DB
}

func (self *pqFacade) MigrateAll() error {
	// TODO: this
	return nil
}

func (self *pqFacade) GetThreads(limit int) ([]model.ThreadInfo, error) {
	// TODO: this
	return nil, nil
}

func NewPQ(conf Config) (*pqFacade, error) {
	db, err := sql.Open("postgres", conf.URL)
	if err != nil {
		return nil, err
	}
	return &pqFacade{db: db}, nil
}
