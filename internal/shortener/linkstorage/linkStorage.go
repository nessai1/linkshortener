package linkstorage

import "fmt"

type HashToLink map[string]string

type StorageDriver interface {
	Save(hl HashToLink) error
	Load() (HashToLink, error)
	Close() error
}

type Storage struct {
	driver     StorageDriver
	hashToLink HashToLink
}

func (storage *Storage) Set(key string, val string) {
	storage.hashToLink[key] = val
}

func (storage *Storage) Get(key string) (string, bool) {
	val, ok := storage.hashToLink[key]
	return val, ok
}

func CreateStorage(driver StorageDriver) (*Storage, error) {
	storage := Storage{
		driver: driver,
	}

	hl, err := driver.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load data from driver: %s", err.Error())
	}

	storage.hashToLink = hl
	return &storage, nil
}
