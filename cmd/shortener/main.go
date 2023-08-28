package main

import (
	"flag"
	"fmt"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
	"github.com/nessai1/linkshortener/internal/storage"
	"os"
)

func initConfig() *shortener.Config {
	serverAddr := flag.String("a", "", "Address of application")
	tokenTail := flag.String("b", "", "Left tail of token of shorted URL")
	storageFilePath := flag.String("f", "./tmp/short-url-db.json", "Path to file storage")

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

	config := shortener.Config{
		ServerAddr:  *serverAddr,
		TokenTail:   *tokenTail,
		StoragePath: *storageFilePath,
	}

	return &config
}

func main() {
	config := initConfig()
	var kvStorage *storage.KeyValueStorage
	var err error
	if config.StoragePath != "" {
		kvStorage, err = storage.GetKVStorage(config.StoragePath)
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
