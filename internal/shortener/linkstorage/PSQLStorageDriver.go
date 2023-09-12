package linkstorage

import (
	"database/sql"
	"fmt"
)

type PSQLStorageDriver struct {
	SQLDriver      *sql.DB
	preparedInsert *sql.Stmt
}

func (driver *PSQLStorageDriver) Set(key string, val string) error {
	_, ok := driver.Get(key)
	if ok {
		return URLIntersectionError
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
	initErr := initTableIfNotExists(driver.SQLDriver)
	if initErr != nil {
		return initErr
	}

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

func initTableIfNotExists(sqlDriver *sql.DB) error {
	if err := sqlDriver.Ping(); err != nil {
		return fmt.Errorf("cannot check table exists, no DB ping: %s", err.Error())
	}

	_, err := sqlDriver.Exec("CREATE TABLE IF NOT EXISTS hash_link (ID serial PRIMARY KEY, HASH varchar(255) NOT NULL UNIQUE, LINK varchar(255) NOT NULL)")
	if err != nil {
		return fmt.Errorf("cannot create database: %s", err.Error())
	}

	return nil
}
