package linkstorage

import "database/sql"

type PSQLStorageDriver struct {
	SQLDriver *sql.DB
}

func (driver *PSQLStorageDriver) Save(hl HashToLink) error {
	return nil
}

func (driver *PSQLStorageDriver) Load() (HashToLink, error) {
	hl := make(HashToLink, 0)
	return hl, nil
}

func (driver *PSQLStorageDriver) Close() error {
	return nil
}
