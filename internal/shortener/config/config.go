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

	flag.Parse()

	config := Config{
		ServerAddr: *serverAddr,
	}

	return &config
}

type Config struct {
	ServerAddr string
	HashTail   string
}
