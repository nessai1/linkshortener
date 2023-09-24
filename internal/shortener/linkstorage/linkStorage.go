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
	Set(key string, val string) error
	Get(key string) (string, bool)
	Ping() (bool, error)
	LoadBatch([]KeyValueRow) error

	Save() error
	Load() error
	Close() error
}

type Storage struct {
	driver StorageDriver
}

func (storage *Storage) Set(link string, val string) error {
	return storage.driver.Set(link, val)
}

func (storage *Storage) Get(link string) (string, bool) {
	val, ok := storage.driver.Get(link)
	return val, ok
}

func (storage *Storage) Save() error {
	return storage.driver.Save()
}

func (storage *Storage) Ping() (bool, error) {
	return storage.driver.Ping()
}

func (storage *Storage) LoadBatch(items []KeyValueRow) error {
	return storage.driver.LoadBatch(items)
}

func CreateStorage(driver StorageDriver) (*Storage, error) {
	storage := Storage{
		driver: driver,
	}

	err := storage.driver.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load data from driver: %s", err.Error())
	}

	return &storage, nil
}
