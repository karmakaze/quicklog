package config

import (
	"flag"
	"os"
	"strconv"
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
	port := 8124
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil && 0 <= p && p < 65536 {
		port = int(p)
	}
	flag.StringVar(&config.Address, "host", "0.0.0.0", "listen tcp address")
	flag.IntVar(&config.Port, "port", port, "listen tcp port (default 443 with SSL)")
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
