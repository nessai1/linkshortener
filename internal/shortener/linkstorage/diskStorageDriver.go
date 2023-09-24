package linkstorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type DiskStorageDriver struct {
	Path string
	hl   HashToLink
}

func (driver *DiskStorageDriver) Set(key string, val string) error {
	_, ok := driver.hl[key]
	if ok {
		return ErrURLIntersection
	}

	driver.hl[key] = val

	return nil
}

func (driver *DiskStorageDriver) Get(key string) (string, bool) {
	val, ok := driver.hl[key]
	return val, ok
}

func (driver *DiskStorageDriver) Save() error {
	kvstruct := make(keyValueStruct, 0)
	for key, val := range driver.hl {
		if key == "" || val == "" {
			continue
		}
		kvstruct = append(kvstruct, KeyValueRow{
			Key:   key,
			Value: val,
		})
	}

	str, err := json.Marshal(&kvstruct)
	if err != nil {
		return err
	}

	file, err := openFile(driver.Path, true)
	if err != nil {
		return err
	}

	_, err = file.Write(str)
	if err == nil {
		return err
	}

	return file.Close()
}

func (driver *DiskStorageDriver) Load() error {
	driver.hl = make(map[string]string, 0)

	file, err := openFile(driver.Path, false)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	nBytes, err := buffer.ReadFrom(file)
	if err != nil {
		return err
	}

	if nBytes == 0 {
		return nil
	}

	var kvsource keyValueStruct
	err = json.Unmarshal(buffer.Bytes(), &kvsource)
	if err != nil {
		return fmt.Errorf("error while unmarshal source: %s", err.Error())
	}

	for _, val := range kvsource {
		driver.hl[val.Key] = val.Value
	}

	return file.Close()
}

func (driver *DiskStorageDriver) Close() error {
	return driver.Save()
}

func (driver *DiskStorageDriver) Ping() (bool, error) {
	return true, nil
}

func (driver *DiskStorageDriver) LoadBatch(items []KeyValueRow) error {
	hasIntersections := false
	for _, item := range items {
		_, ok := driver.hl[item.Key]
		if ok && !hasIntersections {
			hasIntersections = true
		} else {
			driver.hl[item.Key] = item.Value
		}
	}

	if hasIntersections {
		return ErrURLIntersection
	}

	return nil
}

func openFile(path string, write bool) (*os.File, error) {
	dirPath := filepath.Dir(path)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("cannot create dir: %s", err.Error())
	}

	var flags int
	if write {
		flags = os.O_WRONLY | os.O_TRUNC
	} else {
		flags = os.O_RDONLY | os.O_CREATE | os.O_APPEND
	}

	file, err := os.OpenFile(path, flags, 0666)
	if err != nil {
		return nil, err
	}

	return file, nil
}
