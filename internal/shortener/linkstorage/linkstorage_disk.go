package linkstorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DiskLinkStorage структура, реализующая интерфейс LinkStorage через использование файловой системы
type DiskLinkStorage struct {
	MemoryLinkStorage
	filePath string
}

// BeforeShutdown выполняет сохранение данных о ссылках в файл в JSON-формате
func (storage *DiskLinkStorage) BeforeShutdown() error {
	return storage.save()
}

func (storage *DiskLinkStorage) save() error {
	kvstruct := make(keyValueStruct, 0)
	for key, val := range storage.hl {
		if key == "" || val.Value == "" {
			continue
		}
		kvstruct = append(kvstruct, KeyValueRow{
			Key:       key,
			Value:     val.Value,
			OwnerUUID: val.OwnerUUID,
			IsDeleted: val.IsDeleted,
		})
	}

	str, err := json.Marshal(&kvstruct)
	if err != nil {
		return err
	}

	file, err := openFile(storage.filePath, true)
	if err != nil {
		return err
	}

	_, err = file.Write(str)
	if err == nil {
		return err
	}

	return file.Close()
}

// NewFileStorage создает экземпляр хранилища DiskLinkStorage записывая и считывания данные из указанного файла filePath
func NewFileStorage(filePath string) (*DiskLinkStorage, error) {
	storage := DiskLinkStorage{
		filePath: filePath,
	}

	hl, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error while create new file storage: %w", err)
	}

	storage.hl = hl

	return &storage, nil
}

func readFile(filePath string) (HashToLink, error) {
	hl := make(HashToLink)

	file, err := openFile(filePath, false)
	if err != nil {
		return nil, fmt.Errorf("error while read file (open): %w", err)
	}

	var buffer bytes.Buffer
	nBytes, err := buffer.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("error while read file (read bytes): %w", err)
	}

	if nBytes == 0 {
		return hl, nil
	}

	var kvsource keyValueStruct
	err = json.Unmarshal(buffer.Bytes(), &kvsource)
	if err != nil {
		return nil, fmt.Errorf("error while read file (unmarshal): %s", err.Error())
	}

	for _, val := range kvsource {
		hl[val.Key] = Link{
			Value:     val.Value,
			OwnerUUID: val.OwnerUUID,
		}
	}

	return hl, file.Close()
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
