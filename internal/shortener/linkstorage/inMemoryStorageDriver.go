package linkstorage

type InMemoryStorageDriver struct{}

func (driver *InMemoryStorageDriver) Save(hl HashToLink) error {
	return nil
}

func (driver *InMemoryStorageDriver) Load() (HashToLink, error) {
	hl := make(HashToLink, 0)
	return hl, nil
}

func (driver *InMemoryStorageDriver) Close() error {
	return nil
}
