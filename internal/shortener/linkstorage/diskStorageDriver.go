package linkstorage

type DiskStorageDriver struct {
	Path string
}

func (driver *DiskStorageDriver) Save(hl HashToLink) error {
	return nil
}

func (driver *DiskStorageDriver) Load() (HashToLink, error) {
	hl := make(HashToLink, 0)
	return hl, nil
}

func (driver *DiskStorageDriver) Close() error {
	return nil
}
