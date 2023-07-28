package main

import (
	"flag"
	"github.com/nessai1/linkshortener/internal/app"
	"github.com/nessai1/linkshortener/internal/shortener"
	"os"
)

func initConfig() *shortener.Config {
	serverAddr := flag.String("a", "", "Address of application")
	tokenTail := flag.String("b", "", "Left tail of token of shorted URL")

	flag.Parse()

	if serverAddrEnv := os.Getenv("SERVER_ADDRESS"); serverAddrEnv != "" {
		*serverAddr = serverAddrEnv
	}

	if tokenTailEnv := os.Getenv("BASE_URL"); tokenTailEnv != "" {
		*tokenTail = tokenTailEnv
	}

	config := shortener.Config{
		ServerAddr: *serverAddr,
		TokenTail:  *tokenTail,
	}

	return &config
}

func main() {
	app.Run(shortener.GetApplication(initConfig()), app.Stage)
}
