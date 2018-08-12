package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/karmakaze/quicklog/config"
	"github.com/karmakaze/quicklog/web"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	cfg := config.Config{DbUrl: "postgres://postgres:postgres@localhost:5432/quicklog?sslmode=disable"}

	err := web.Serve("0.0.0.0", cfg)
	if err != nil {
		fmt.Printf("Error serving 0.0.0.0: %v\n", err)
	}
}
