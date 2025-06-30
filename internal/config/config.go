package config

import "flag"

var FlagRunAddr string
var FlagBaseAddr string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "Run server address")
	flag.StringVar(&FlagBaseAddr, "b", "http://localhost:8080", "Base server address")
	flag.Parse()
}
