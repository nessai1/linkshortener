package linkstorage

import "fmt"

type HashToLink map[string]string

type StorageDriver interface {
	Set(key string, val string)
	Get(key string) (string, bool)

	Save() error
	Load() error
	Close() error
}

type Storage struct {
	driver StorageDriver
}

func (storage *Storage) Set(key string, val string) {
	storage.driver.Set(key, val)
}

func (storage *Storage) Get(key string) (string, bool) {
	val, ok := storage.driver.Get(key)
	return val, ok
}

func (storage *Storage) Save() error {
	return storage.driver.Save()
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
