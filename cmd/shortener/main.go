package main

import (
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
)

func main() {

	shortenerConfig, err := shortener.BuildAppConfig()

	if err != nil {
		panic(err)
	}

	app.Run(shortener.GetApplication(shortenerConfig), app.Development)
}
