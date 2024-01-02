package shortener

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"

	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// BuildAppConfig собирает конфигурацию для shortener приложения исходя из входящей конфигурации ENV&flags
func BuildAppConfig() (*Config, error) {
	config := fetchConfig()
	linkStorage, err := chooseLinkStorage(&config)
	if err != nil {
		return nil, fmt.Errorf("cannot build app config: %w", err)
	}

	shortenerConfig := Config{
		ServerAddr:  config.ServerAddr,
		TokenTail:   config.TokenTail,
		LinkStorage: linkStorage,
	}

	return &shortenerConfig, nil
}

// InitConfig сырые конфигурационные данные сервера
type InitConfig struct {
	// ServerAddr адрес сервера
	ServerAddr string
	// TokenTail префикс с которым будет возвращаться результат хеширования ссылки
	TokenTail string
	// SQLConnection строка с настройками соединения к СУБД
	SQLConnection string
	// FileStoragePath путь файла в который будет записывать файловый репозиторий ссылок
	FileStoragePath string
}

func fetchConfig() InitConfig {
	serverAddr := flag.String("a", "", "Address of application")
	tokenTail := flag.String("b", "", "Left tail of token of shorted URL")
	storageFilePath := flag.String("f", "./tmp/short-url-db.json", "Path to file storage")
	postgresConnParams := flag.String("d", "", "Connection params for postgres")

	flag.Parse()

	if serverAddrEnv := os.Getenv("SERVER_ADDRESS"); serverAddrEnv != "" {
		*serverAddr = serverAddrEnv
	}

	if tokenTailEnv := os.Getenv("BASE_URL"); tokenTailEnv != "" {
		*tokenTail = tokenTailEnv
	}

	if storageFilePathEnv := os.Getenv("FILE_STORAGE_PATH"); storageFilePathEnv != "" {
		*storageFilePath = storageFilePathEnv
	}

	if postgresConnParamsEnv := os.Getenv("DATABASE_DSN"); postgresConnParamsEnv != "" {
		*postgresConnParams = postgresConnParamsEnv
	}

	return InitConfig{
		ServerAddr:      *serverAddr,
		TokenTail:       *tokenTail,
		SQLConnection:   *postgresConnParams,
		FileStoragePath: *storageFilePath,
	}
}

func initMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	migrations, err := migrate.NewWithDatabaseInstance("file:migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("error while create migrations: %s", err.Error())
	}

	if err = migrations.Up(); err != nil {
		return err
	}

	return nil
}

func chooseLinkStorage(config *InitConfig) (linkstorage.LinkStorage, error) {
	if config.SQLConnection != "" {
		db, err := sql.Open("pgx", config.SQLConnection)
		if err != nil {
			return nil, err
		}

		err = initMigrations(db)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return nil, err
		}

		linkStorage, err := linkstorage.NewPsqlStorage(db)
		if err != nil {
			return nil, fmt.Errorf("cannot choose psql driver: %w", err)
		}

		return linkStorage, nil
	} else if config.FileStoragePath != "" {
		linkStorage, err := linkstorage.NewFileStorage(config.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("cannot choose file driver: %w", err)
		}

		return linkStorage, nil
	}

	return linkstorage.NewMemoryStorage(nil), nil
}
