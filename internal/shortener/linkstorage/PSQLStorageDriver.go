package linkstorage

import (
	"database/sql"
)

type PSQLStorageDriver struct {
	SQLDriver *sql.DB
}

func (driver *PSQLStorageDriver) Save(hl HashToLink) error {
	preparedInsert, prepareErr := driver.SQLDriver.Prepare("INSERT INTO hash_link (HASH, LINK) VALUES ($1, $2) ON CONFLICT DO NOTHING")
	if prepareErr != nil {
		return prepareErr
	}

	tx, err := driver.SQLDriver.Begin()
	if err != nil {
		return err
	}

	for key, val := range hl {
		_, err = preparedInsert.Exec(key, val)
		if err != nil {
			return tx.Rollback()
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return preparedInsert.Close()
}

func (driver *PSQLStorageDriver) Load() (HashToLink, error) {
	hl := make(HashToLink, 0)

	rows, err := driver.SQLDriver.Query("SELECT hash, link FROM hash_link")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	kvrow := KeyValueRow{}
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}
		err = rows.Scan(&kvrow.Key, &kvrow.Value)
		if err != nil {
			return nil, err
		}

		hl[kvrow.Key] = kvrow.Value
	}

	return hl, nil
}

func (driver *PSQLStorageDriver) Close() error {
	return driver.SQLDriver.Close()
}

func (driver *PSQLStorageDriver) Ping() (bool, error) {
	connectionStatus := driver.SQLDriver.Ping()
	if connectionStatus == nil {
		return true, nil
	}

	return false, connectionStatus
}
