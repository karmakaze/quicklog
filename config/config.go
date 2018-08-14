package config

import (
	"flag"
)

type Config struct {
	Address      string
	Port         int
	DbUrl        string
	SslFullChain string
	SslPrivKey   string
}

var config Config

func Parse() Config {
	flag.StringVar(&config.Address, "host", "0.0.0.0", "listen tcp address")
	flag.IntVar(&config.Port, "port", 8124, "listen tcp port (default 443 with SSL)")
	flag.StringVar(&config.DbUrl, "dbUrl", "postgres://quicklog:quicklog@localhost:5432/quicklog?sslmode=disable",
		"database connection url")
	flag.StringVar(&config.SslFullChain, "sslCertFile", "", "path/name of fullchain.pem (optional)")
	flag.StringVar(&config.SslPrivKey, "sslPrivKey", "", "path/name of privkey.pem (optional)")

	flag.Parse()

	if config.SslFullChain != "" && config.SslPrivKey != "" && config.Port == 8124 {
		config.Port = 443
	}
	return config
}
