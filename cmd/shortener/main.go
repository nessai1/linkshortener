package main

import (
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
)

func main() {
	app.Run(shortener.GetApplication())
}
