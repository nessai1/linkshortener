package linkstorage

type InMemoryStorageDriver struct {
	hl HashToLink
}

func (driver *InMemoryStorageDriver) Set(key string, val string) error {
	_, ok := driver.hl[key]
	if ok {
		return ErrURLIntersection
	}

	driver.hl[key] = val

	return nil
}

func (driver *InMemoryStorageDriver) Get(key string) (string, bool) {
	val, ok := driver.hl[key]
	return val, ok
}

func (driver *InMemoryStorageDriver) Save() error {
	return nil
}

func (driver *InMemoryStorageDriver) Load() error {
	driver.hl = make(HashToLink, 0)
	return nil
}

func (driver *InMemoryStorageDriver) Close() error {
	return nil
}
