package linkstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nessai1/linkshortener/internal/postgrescodes"
)

type PsqlLinkStorage struct {
	db            *sql.DB
	insertCommand *sql.Stmt
}

func (storage *PsqlLinkStorage) Set(ctx context.Context, hash string, link Link) error {
	_, err := storage.insertCommand.ExecContext(ctx, hash, link.Value, link.OwnerUUID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == postgrescodes.PostgresErrCodeUniqueViolation {
				return ErrURLIntersection
			}
		}

		return fmt.Errorf("[psql storage] error while set URL: %w", err)
	}

	return nil
}

func (storage *PsqlLinkStorage) Get(ctx context.Context, hash string) (Link, bool) {
	link := Link{}

	err := storage.db.QueryRowContext(
		ctx,
		"SELECT link, owner_uuid, is_deleted FROM hash_link WHERE hash = $1",
		hash,
	).Scan(&link.Value, &link.OwnerUUID, &link.IsDeleted)

	if err != nil {
		return Link{}, false
	}

	return link, true
}

// TODO: need to return error if some shit happened while query
func (storage *PsqlLinkStorage) FindByUserUUID(ctx context.Context, userUUID string) []KeyValueRow {
	rows, err := storage.db.QueryContext(
		ctx,
		"SELECT hash, link, owner_uuid, is_deleted FROM hash_link WHERE owner_uuid = $1",
		userUUID,
	)

	if err != nil {
		return nil
	}

	resultRows := make([]KeyValueRow, 0)
	for rows.Next() {
		if err = rows.Err(); err != nil {
			continue
		}

		kvrow := KeyValueRow{}
		err = rows.Scan(&kvrow.Key, &kvrow.Value, &kvrow.OwnerUUID, &kvrow.IsDeleted)
		if err != nil {
			continue
		}

		resultRows = append(resultRows, kvrow)
	}

	return resultRows
}

func (storage *PsqlLinkStorage) Ping() (bool, error) {
	err := storage.db.Ping()
	return err == nil, err
}

func (storage *PsqlLinkStorage) LoadBatch(ctx context.Context, items []KeyValueRow) error {
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("[psql storage] error while load batch (start transaction): %w", err)
	}

	for _, item := range items {
		link := Link{
			Value:     item.Value,
			OwnerUUID: item.OwnerUUID,
			IsDeleted: item.IsDeleted,
		}

		if err = storage.Set(ctx, item.Key, link); err != nil && !errors.Is(err, ErrURLIntersection) {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("[psql storage] error while load link batch (rollback&set error): %w", errors.Join(rollbackErr, err))
			} else {
				return fmt.Errorf("[psql storage] error while load link batch (set error): %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("[psql storage] error while load link batch (commit): %w", err)
	}

	return nil
}

func (storage *PsqlLinkStorage) DeleteBatch(ctx context.Context, items []Hash) error {
	preparedDelete, err := storage.db.PrepareContext(ctx, "DELETE FROM hash_link WHERE hash = $1 AND owner_uuid = $2")
	if err != nil {
		return fmt.Errorf("[psql storage] error while delete batch (prepare command): %w", err)
	}

	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("[psql storage] error while delete batch (start transaction): %w", err)
	}

	for _, item := range items {
		_, err = preparedDelete.ExecContext(ctx, item.Value, item.OwnerUUID)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("[psql storage] error while delete link batch (rollback&set error): %w", errors.Join(rollbackErr, err))
			} else {
				return fmt.Errorf("[psql storage] error while delete link batch (set error): %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("[psql storage] error while delete link batch (commit): %w", err)
	}

	return nil
}

func (storage *PsqlLinkStorage) BeforeShutdown() error {
	return storage.db.Close()
}

func NewPsqlStorage(db *sql.DB) (*PsqlLinkStorage, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot ping db while create psql storage: %w", err)
	}

	insertCommand, err := db.Prepare("INSERT INTO hash_link (hash, link, owner_uuid) VALUES ($1, $2, $3)")
	if err != nil {
		return nil, fmt.Errorf("cannot prepare insert command while create psql storage: %w", err)
	}

	return &PsqlLinkStorage{db: db, insertCommand: insertCommand}, nil
}
