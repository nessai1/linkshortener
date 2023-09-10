package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
	"github.com/nessai1/linkshortener/internal/storage"
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

	config := shortener.Config{
		ServerAddr:  *serverAddr,
		TokenTail:   *tokenTail,
		StoragePath: *storageFilePath,
		SQLDriver:   db,
	}

	return &config
}

func main() {
	config := initConfig()
	defer config.SQLDriver.Close()

	var kvStorage *storage.KeyValueStorage
	var err error
	if config.StoragePath != "" {
		kvStorage, err = storage.GetFileKVStorage(config.StoragePath)
		if err != nil {
			panic(fmt.Sprintf("Cannot open storage on path %s! Application is shutting down. (%s)", config.StoragePath, err.Error()))
		}
	} else {
		kvStorage, err = storage.CreateTempKVStorage()
		if err != nil {
			panic(fmt.Sprintf("Cannot create temp storage! Application is shutting down. (%s)", err.Error()))
		}
	}

	app.Run(shortener.GetApplication(config, kvStorage), app.Stage)
}
