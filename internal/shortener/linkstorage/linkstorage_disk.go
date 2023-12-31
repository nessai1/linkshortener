package linkstorage

import (
	"context"
)

type DiskLinkStorage struct {
}

func (storage *DiskLinkStorage) Set(ctx context.Context, hash string, link Link) error {
	//TODO implement me
	panic("implement me")
}

func (storage *DiskLinkStorage) Get(ctx context.Context, hash string) (Link, bool) {
	//TODO implement me
	panic("implement me")
}

func (storage *DiskLinkStorage) FindByUserUUID(ctx context.Context, userUUID string) []KeyValueRow {
	//TODO implement me
	panic("implement me")
}

func (storage *DiskLinkStorage) Ping() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (storage *DiskLinkStorage) LoadBatch(ctx context.Context, items []KeyValueRow) error {
	//TODO implement me
	panic("implement me")
}

func (storage *DiskLinkStorage) DeleteBatch(ctx context.Context, items []Hash) error {
	//TODO implement me
	panic("implement me")
}

func (storage *DiskLinkStorage) BeforeShutdown() error {
	//TODO implement me
	panic("implement me")
}

func NewFileStorage(filePath string) (*DiskLinkStorage, error) {
	return nil, nil // TODO: implement me
}
