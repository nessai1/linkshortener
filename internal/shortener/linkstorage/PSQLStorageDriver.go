package linkstorage

import (
	"database/sql"
)

type PSQLStorageDriver struct {
	SQLDriver      *sql.DB
	preparedInsert *sql.Stmt
}

func (driver *PSQLStorageDriver) Set(key string, val string) error {
	_, ok := driver.Get(key)
	if ok {
		return ErrURLIntersection
	}

	_, err := driver.preparedInsert.Exec(key, val)
	return err
}

func (driver *PSQLStorageDriver) Get(key string) (string, bool) {
	row := driver.SQLDriver.QueryRow("SELECT LINK FROM hash_link WHERE HASH = $1", key)

	var hash string
	if row.Scan(&hash) == sql.ErrNoRows {
		return "", false
	}

	return hash, true
}

func (driver *PSQLStorageDriver) Save() error {
	return nil
}

func (driver *PSQLStorageDriver) Load() error {
	var prepareErr error
	driver.preparedInsert, prepareErr = driver.SQLDriver.Prepare("INSERT INTO hash_link (HASH, LINK) VALUES ($1, $2)")
	if prepareErr != nil {
		return prepareErr
	}

	return nil
}

func (driver *PSQLStorageDriver) Close() error {
	err := driver.preparedInsert.Close()
	if err != nil {
		return err
	}

	return driver.SQLDriver.Close()
}

func (driver *PSQLStorageDriver) Ping() (bool, error) {
	connectionStatus := driver.SQLDriver.Ping()
	if connectionStatus == nil {
		return true, nil
	}

	return false, connectionStatus
}

func (driver *PSQLStorageDriver) LoadBatch(batch []KeyValueRow) error {
	tx, err := driver.SQLDriver.Begin()
	if err != nil {
		return err
	}

	for _, item := range batch {
		err := driver.Set(item.Key, item.Value)
		if err != nil {
			return tx.Rollback()
		}
	}

	return tx.Commit()
}
