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
}

func (driver *DiskStorageDriver) Save(hl HashToLink) error {
	kvstruct := make(keyValueStruct, 0)
	for key, val := range hl {
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

func (driver *DiskStorageDriver) Load() (HashToLink, error) {
	hl := make(HashToLink, 0)

	file, err := openFile(driver.Path, false)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	nBytes, err := buffer.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	if nBytes == 0 {
		return hl, nil
	}

	var kvsource keyValueStruct
	err = json.Unmarshal(buffer.Bytes(), &kvsource)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshal source: %s", err.Error())
	}

	for _, val := range kvsource {
		hl[val.Key] = val.Value
	}

	return hl, file.Close()
}

func (driver *DiskStorageDriver) Close() error {
	return nil
}

func (driver *DiskStorageDriver) Ping() (bool, error) {
	return true, nil
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
