package linkstorage

import (
	"context"
	"errors"
)

// HashToLink описывает карту хеш -> данные о ссылке
type HashToLink map[string]Link

// Link структура описывающая ссылку
type Link struct {
	// Value непосредственно ссылка
	Value string
	// OwnerUUID уникальный идентификатор пользователя, создавшего ссылку
	OwnerUUID string
	// IsDeleted флаг, маркирующий удалена ли ссылка или нет
	IsDeleted bool
}

// Hash структура описывающая хеш от ссылки
type Hash struct {
	// Value непосредственно хеш
	Value string
	// OwnerUUID уникальный идентификатор пользователя, создавшего хеш
	OwnerUUID string
}

// KeyValueRow структура описывающая связь хеш -> ссылка
type KeyValueRow struct {
	// Key значение хеша
	Key string `json:"key"`
	// Value значение ссылки
	Value string `json:"value"`
	// OwnerUUID владелец связки
	OwnerUUID string `json:"owner_uuid"`
	// IsDeleted маркер удалена ли ссылка
	IsDeleted bool `json:"is_deleted"`
}

// ErrURLIntersection ошибка, возникающая при попытке добавить существующую ссылку
var ErrURLIntersection = errors.New("inserting URL not unique")

type keyValueStruct []KeyValueRow

// LinkStorage интерфейс, описывающий репозиторий для работы с ссылками
type LinkStorage interface {
	// Set добавляет в репозиторий новую связку хеш - ссылка
	Set(ctx context.Context, hash string, link Link) error
	// Get получает ссылку по указанному хешу, если ссылка не найдена - второй аргумент == false
	Get(ctx context.Context, hash string) (Link, bool)
	// FindByUserUUID выполняет поиск и возвращает список ссылок, привязанных к конкретному идентификатору пользователя
	FindByUserUUID(ctx context.Context, userUUID string) []KeyValueRow
	// Ping проверяет, работоспособно ли хранилище. В случае неработоспособности первый аргумент == false, второй содержит ошибку
	Ping() (bool, error)

	// LoadBatch загружает пачку ссылок в хранилище
	LoadBatch(ctx context.Context, items []KeyValueRow) error
	// DeleteBatch маркирует пачку ссылок как удаленные
	DeleteBatch(ctx context.Context, items []Hash) error

	// BeforeShutdown метод, вызывающийся перед закрытием хранилища
	BeforeShutdown() error
}
