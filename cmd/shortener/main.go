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

func initConfig() *shortener.Config {
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

	db, err := sql.Open("pgx", *postgresConnParams)
	if err != nil {
		panic("Cannot create DB driver: " + err.Error())
	}

	var storageDriver linkstorage.StorageDriver

	if *postgresConnParams != "" {
		storageDriver = &linkstorage.PSQLStorageDriver{SQLDriver: db}
		err = initMigrations(db)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			panic(err)
		}
	} else if *storageFilePath != "" {
		storageDriver = &linkstorage.DiskStorageDriver{Path: *storageFilePath}
	} else {
		storageDriver = &linkstorage.InMemoryStorageDriver{}
	}

	config := shortener.Config{
		ServerAddr:    *serverAddr,
		TokenTail:     *tokenTail,
		SQLDriver:     db,
		StorageDriver: storageDriver,
	}

	return &config
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

func main() {
	config := initConfig()

	app.Run(shortener.GetApplication(config), app.Development)
}
