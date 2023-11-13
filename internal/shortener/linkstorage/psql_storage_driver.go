package linkstorage

import (
	"database/sql"
)

type PSQLStorageDriver struct {
	SQLDriver *sql.DB
}

func (driver *PSQLStorageDriver) Save(hl HashToLink) error {
	preparedInsert, prepareErr := driver.SQLDriver.Prepare("INSERT INTO hash_link (HASH, LINK, OWNER_UUID, IS_DELETED) VALUES ($1, $2, $3, $4) ON CONFLICT (HASH) DO UPDATE SET IS_DELETED = $4")
	if prepareErr != nil {
		return prepareErr
	}

	tx, err := driver.SQLDriver.Begin()
	if err != nil {
		return err
	}

	for key, val := range hl {
		var ownerUUID *string
		if val.OwnerUUID != "" {
			ownerUUID = &val.OwnerUUID
		} else {
			ownerUUID = nil
		}

		_, err = preparedInsert.Exec(key, val.Value, ownerUUID, val.IsDeleted)
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

	rows, err := driver.SQLDriver.Query("SELECT HASH, LINK, OWNER_UUID FROM hash_link")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	kvrow := KeyValueRow{}
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}

		var ownerUUID *string
		err = rows.Scan(&kvrow.Key, &kvrow.Value, &ownerUUID)
		if err != nil {
			return nil, err
		}

		if ownerUUID == nil {
			hl[kvrow.Key] = Link{
				Value:     kvrow.Value,
				OwnerUUID: "",
			}
		} else {
			hl[kvrow.Key] = Link{
				Value:     kvrow.Value,
				OwnerUUID: *ownerUUID,
			}
		}

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
