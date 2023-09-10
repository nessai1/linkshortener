package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type KeyValueStorage struct {
	physicalStorage io.ReadWriteCloser
	keyValueMap     map[string]string

	isTemp     bool
	storageDir string
}

type keyValueRow struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type keyValueStruct []keyValueRow

func (storage *KeyValueStorage) Set(key string, val string) {
	storage.keyValueMap[key] = val
}

func (storage *KeyValueStorage) Get(key string) (string, bool) {
	val, ok := storage.keyValueMap[key]
	return val, ok
}

func (storage *KeyValueStorage) Save() error {

	kvstruct := make(keyValueStruct, 0)
	for key, val := range storage.keyValueMap {
		if key == "" || val == "" {
			continue
		}
		kvstruct = append(kvstruct, keyValueRow{
			Key:   key,
			Value: val,
		})
	}

	str, err := json.Marshal(&kvstruct)
	if err != nil {
		return err
	}

	_, err = storage.physicalStorage.Write(str)
	if err == nil {
		return err
	}

	return nil
}

func (storage *KeyValueStorage) Close() error {
	err := storage.Save()
	if err != nil {
		return err
	}

	return storage.physicalStorage.Close()
}

func GetFileKVStorage(path string) (*KeyValueStorage, error) {
	dirPath := filepath.Dir(path)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("cannot create dir: %s", err.Error())
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	kvmap, err := fillMap(file)
	if err != nil {
		return nil, err
	}
	file.Truncate(0)
	file.Seek(0, 0)

	file, err = os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &KeyValueStorage{
		physicalStorage: file,
		keyValueMap:     kvmap,
		isTemp:          false,
		storageDir:      dirPath,
	}, nil
}

func fillMap(rwc io.ReadWriteCloser) (map[string]string, error) {
	var buffer bytes.Buffer
	nBytes, err := buffer.ReadFrom(rwc)
	if err != nil {
		return nil, err
	}
	if nBytes == 0 {
		kvsource := make(map[string]string, 10)
		return kvsource, nil
	}

	var kvsource keyValueStruct
	err = json.Unmarshal(buffer.Bytes(), &kvsource)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshal source: %s", err.Error())
	}

	kvmap := make(map[string]string, len(kvsource))

	for _, val := range kvsource {
		kvmap[val.Key] = val.Value
	}

	return kvmap, nil
}

/*
Создание файлов для временных хранилищ планировалось для того,
что-бы в будущем при цикличной синхронизации файла с мапой, раз в N времени, можно было мониторить что там внутри храниться
(если конечно руки дойдут до этой фичи)
*/
func CreateTempKVStorage() (*KeyValueStorage, error) {
	path, err := os.MkdirTemp(".", "storage")
	if err != nil {
		return nil, fmt.Errorf("cannot create dir: %s", err.Error())
	}

	file, err := os.CreateTemp(path, "tempStorageFile*")
	if err != nil {
		return nil, err
	}

	kvmap := make(map[string]string, 10)

	return &KeyValueStorage{
		physicalStorage: file,
		keyValueMap:     kvmap,
		isTemp:          true,
	}, nil
}
