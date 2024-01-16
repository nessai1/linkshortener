package linkstorage

import (
	"context"
	"errors"
)

// MemoryLinkStorage структура, реализующая интерфейс LinkStorage через использование структуры map
type MemoryLinkStorage struct {
	hl HashToLink
}

// Set сохраняет ссылку в карте map[string]Link по ключу hash
func (storage *MemoryLinkStorage) Set(_ context.Context, hash string, link Link) error {
	_, ok := storage.hl[hash]
	if !ok {
		storage.hl[hash] = link
		return nil
	}

	return ErrURLIntersection
}

// Get получает ссылку из карты map[string]Link по ключу hash
func (storage *MemoryLinkStorage) Get(_ context.Context, hash string) (Link, bool) {
	val, ok := storage.hl[hash]
	return val, ok
}

// FindByUserUUID ищет среди ссылок в карте map[string]Link ссылки с OwnerUUID === userUUID
func (storage *MemoryLinkStorage) FindByUserUUID(_ context.Context, userUUID string) []KeyValueRow {
	rows := make([]KeyValueRow, 0)

	for key, val := range storage.hl {
		if val.OwnerUUID == userUUID {
			klr := KeyValueRow{
				Key:       key,
				Value:     val.Value,
				OwnerUUID: val.OwnerUUID,
				IsDeleted: val.IsDeleted,
			}

			rows = append(rows, klr)
		}
	}

	return rows
}

// Ping проверка доступности карты map[string]Link. Всегда true
func (storage *MemoryLinkStorage) Ping() (bool, error) {
	return true, nil
}

// LoadBatch загружает ссылки указанные в items в карту map[string]Link
func (storage *MemoryLinkStorage) LoadBatch(_ context.Context, items []KeyValueRow) error {
	for _, item := range items {
		link := Link{
			Value:     item.Value,
			OwnerUUID: item.OwnerUUID,
			IsDeleted: false,
		}

		err := storage.Set(context.TODO(), item.Key, link)
		if err != nil && !errors.Is(err, ErrURLIntersection) {
			return err
		}
	}

	return nil
}

// DeleteBatch удаляет из карты map[string]Link пачку ссылок, совпадающих по Hash.Value && Hash.OwnerUUID
func (storage *MemoryLinkStorage) DeleteBatch(_ context.Context, items []Hash) error {
	for _, item := range items {
		val := storage.hl[item.Value]
		if val.OwnerUUID == item.OwnerUUID {
			val.IsDeleted = true
			storage.hl[item.Value] = val
		}
	}

	return nil
}

// BeforeShutdown не делает никаких операций перед закрытием приложения
func (storage *MemoryLinkStorage) BeforeShutdown() error {
	return nil
}

// NewMemoryStorage создает экземпляр хранилища MemoryLinkStorage, в случае указания initData карта будет заполнена первоначальными данными из этой карты
func NewMemoryStorage(initData HashToLink) *MemoryLinkStorage {
	var hl HashToLink

	if initData != nil {
		hl = initData
	} else {
		hl = make(HashToLink)
	}

	return &MemoryLinkStorage{hl: hl}
}
