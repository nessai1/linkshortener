package main

import (
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {

	shortenerConfig, err := shortener.BuildAppConfig()

	if err != nil {
		panic(err)
	}

	app.Run(shortener.GetApplication(shortenerConfig), app.Development, getApplicationInfo())
}

func getApplicationInfo() app.ApplicationInfo {
	info := app.ApplicationInfo{
		BuildVersion: buildVersion,
		BuildDate:    buildDate,
		BuildCommit:  buildCommit,
	}

	return info
}
