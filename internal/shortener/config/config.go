package config

import "flag"

var configInstance *Config

func GetConfig() *Config {
	if configInstance == nil {
		configInstance = initConfig()
	}

	return configInstance
}

func initConfig() *Config {
	serverAddr := flag.String("a", "", "Address of application")
	tokenTail := flag.String("b", "", "Left tail of token of shorted URL")

	flag.Parse()

	config := Config{
		ServerAddr: *serverAddr,
		TokenTail:  *tokenTail,
	}

	return &config
}

type Config struct {
	ServerAddr string
	TokenTail  string
}
