package main

import (
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
)

var buildVersion string
var buildDate string
var buildCommit string

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

	if info.BuildVersion == "" {
		info.BuildVersion = "N/A"
	}

	if info.BuildDate == "" {
		info.BuildDate = "N/A"
	}

	if info.BuildCommit == "" {
		info.BuildCommit = "N/A"
	}

	return info
}
