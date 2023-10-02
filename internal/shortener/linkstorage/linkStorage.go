package linkstorage

import (
	"errors"
	"fmt"
)

type HashToLink map[string]string

type KeyValueRow struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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

func (storage *Storage) Set(link string, val string) error {
	_, ok := storage.hl[link]
	if ok {
		return ErrURLIntersection
	}
	storage.hl[link] = val
	if storage.driver != nil {
		return storage.Save()
	}

	return nil
}

func (storage *Storage) Get(link string) (string, bool) {
	val, ok := storage.hl[link]
	return val, ok
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
		storage.hl[item.Key] = item.Value
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
		storage.hl = make(map[string]string, 0)
	}

	return &storage, nil
}
