package linkstorage

import (
	"errors"
	"fmt"
)

type HashToLink map[string]Link

type Link struct {
	Value     string
	OwnerUUID string
}

type KeyValueRow struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	OwnerUUID string `json:"owner_uuid"`
}

var ErrURLIntersection = errors.New("inserting URL not unique")

type keyValueStruct []KeyValueRow

type StorageDriver interface {
	Save(HashToLink) error
	Load() (HashToLink, error)
	Close() error
	Ping() (bool, error)
}

type Storage struct {
	driver StorageDriver
	hl     HashToLink
}

func (storage *Storage) Set(hash string, link Link) error {
	_, ok := storage.hl[hash]
	if ok {
		return ErrURLIntersection
	}
	storage.hl[hash] = link
	if storage.driver != nil {
		return storage.Save()
	}

	return nil
}

func (storage *Storage) Get(hash string) (Link, bool) {
	link, ok := storage.hl[hash]
	return link, ok
}

func (storage *Storage) FindByUserUUID(userUUID string) []KeyValueRow {
	links := make([]KeyValueRow, 0)
	for hash, link := range storage.hl {
		if link.OwnerUUID == userUUID {
			links = append(links, KeyValueRow{
				Key:       hash,
				Value:     link.Value,
				OwnerUUID: link.OwnerUUID,
			})
		}
	}

	return links
}

func (storage *Storage) Save() error {
	if storage.driver != nil {
		return storage.driver.Save(storage.hl)
	}

	return nil
}

func (storage *Storage) Ping() (bool, error) {
	if storage.driver != nil {
		return storage.driver.Ping()
	}

	return true, nil
}

func (storage *Storage) LoadBatch(items []KeyValueRow) error {
	for _, item := range items {
		storage.hl[item.Key] = Link{
			Value:     item.Value,
			OwnerUUID: item.OwnerUUID,
		}
	}

	return nil
}

func CreateStorage(driver StorageDriver) (*Storage, error) {
	storage := Storage{
		driver: driver,
	}

	if driver != nil {
		hl, err := storage.driver.Load()
		if err != nil {
			return nil, fmt.Errorf("cannot load data from driver: %s", err.Error())
		}

		storage.hl = hl
	} else {
		storage.hl = make(map[string]Link, 0)
	}

	return &storage, nil
}
