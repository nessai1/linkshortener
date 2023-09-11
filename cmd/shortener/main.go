package main

import (
	"database/sql"
	"flag"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
	"github.com/nessai1/linkshortener/internal/shortener/linkstorage"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
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

func main() {
	config := initConfig()
	defer config.SQLDriver.Close()

	app.Run(shortener.GetApplication(config), app.Development)
}
