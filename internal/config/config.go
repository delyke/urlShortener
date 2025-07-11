package config

import "flag"

type Config struct {
	RunAddr  string
	BaseAddr string
}

func GetConfig() *Config {
	runAddr := flag.String("a", ":8080", "Run server address")
	baseAddr := flag.String("b", "http://localhost:8080", "Base server address")
	flag.Parse()

	return &Config{
		RunAddr:  *runAddr,
		BaseAddr: *baseAddr,
	}
}
