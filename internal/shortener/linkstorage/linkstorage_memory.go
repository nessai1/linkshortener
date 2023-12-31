package linkstorage

import (
	"context"
	"errors"
)

type MemoryLinkStorage struct {
	hl HashToLink
}

func (storage *MemoryLinkStorage) Set(_ context.Context, hash string, link Link) error {
	_, ok := storage.hl[hash]
	if !ok {
		storage.hl[hash] = link
		return nil
	}

	return ErrURLIntersection
}

func (storage *MemoryLinkStorage) Get(_ context.Context, hash string) (Link, bool) {
	val, ok := storage.hl[hash]
	return val, ok
}

func (storage *MemoryLinkStorage) FindByUserUUID(_ context.Context, userUUID string) []KeyValueRow {
	rows := make([]KeyValueRow, 0)

	for key, val := range storage.hl {
		if val.OwnerUUID == userUUID {
			klr := KeyValueRow{
				Key:       key,
				Value:     val.Value,
				OwnerUUID: val.OwnerUUID,
				IsDeleted: val.IsDeleted,
			}

			rows = append(rows, klr)
		}
	}

	return rows
}

func (storage *MemoryLinkStorage) Ping() (bool, error) {
	return true, nil
}

func (storage *MemoryLinkStorage) LoadBatch(_ context.Context, items []KeyValueRow) error {
	for _, item := range items {
		link := Link{
			Value:     item.Value,
			OwnerUUID: item.OwnerUUID,
			IsDeleted: false,
		}

		err := storage.Set(context.TODO(), item.Key, link)
		if err != nil && !errors.Is(err, ErrURLIntersection) {
			return err
		}
	}

	return nil
}

func (storage *MemoryLinkStorage) DeleteBatch(_ context.Context, items []Hash) error {
	for _, item := range items {
		val := storage.hl[item.Value]
		if val.OwnerUUID == item.OwnerUUID {
			val.IsDeleted = true
			storage.hl[item.Value] = val
		}
	}

	return nil
}

func (storage *MemoryLinkStorage) BeforeShutdown() error {
	return nil
}

func NewMemoryStorage(initData HashToLink) *MemoryLinkStorage {
	var hl HashToLink

	if initData != nil {
		hl = initData
	} else {
		hl = make(HashToLink)
	}

	return &MemoryLinkStorage{hl: hl}
}
