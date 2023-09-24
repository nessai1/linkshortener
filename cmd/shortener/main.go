package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type InitConfig struct {
	ServerAddr      string
	TokenTail       string
	SQLConnection   string
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

func chooseDriver(config *InitConfig) (*linkstorage.StorageDriver, error) {
	var storageDriver linkstorage.StorageDriver

	if config.SQLConnection != "" {
		db, err := sql.Open("pgx", config.SQLConnection)
		if err != nil {
			return nil, err
		}

		storageDriver = &linkstorage.PSQLStorageDriver{SQLDriver: db}
		err = initMigrations(db)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			panic(err)
		}
	} else if config.FileStoragePath != "" {
		storageDriver = &linkstorage.DiskStorageDriver{Path: config.FileStoragePath}
	} else {
		storageDriver = &linkstorage.InMemoryStorageDriver{}
	}

	return &storageDriver, nil
}

func main() {
	config := fetchConfig()
	storageDriver, err := chooseDriver(&config)
	if err != nil {
		panic(err)
	}

	shortenerConfig := shortener.Config{
		ServerAddr:    config.ServerAddr,
		TokenTail:     config.TokenTail,
		StorageDriver: *storageDriver,
	}

	app.Run(shortener.GetApplication(&shortenerConfig), app.Development)
}
