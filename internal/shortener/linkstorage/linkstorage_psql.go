package linkstorage

import (
	"context"
	"database/sql"
	"fmt"
)

type PsqlLinkStorage struct {
	db *sql.DB
}

func (storage *PsqlLinkStorage) Set(ctx context.Context, hash string, link Link) error {
	//TODO implement me
	panic("implement me")
}

func (storage *PsqlLinkStorage) Get(ctx context.Context, hash string) (Link, bool) {
	//TODO implement me
	panic("implement me")
}

func (storage *PsqlLinkStorage) FindByUserUUID(ctx context.Context, userUUID string) []KeyValueRow {
	//TODO implement me
	panic("implement me")
}

func (storage *PsqlLinkStorage) Ping() (bool, error) {
	err := storage.db.Ping()
	return err != nil, err
}

func (storage *PsqlLinkStorage) LoadBatch(ctx context.Context, items []KeyValueRow) error {
	//TODO implement me
	panic("implement me")
}

func (storage *PsqlLinkStorage) DeleteBatch(ctx context.Context, items []Hash) error {
	//TODO implement me
	panic("implement me")
}

func (storage *PsqlLinkStorage) BeforeShutdown() error {
	return storage.db.Close()
}

func NewPsqlStorage(db *sql.DB) (*PsqlLinkStorage, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot ping db while create psql driver: %w", err)
	}

	return &PsqlLinkStorage{db: db}, nil
}
