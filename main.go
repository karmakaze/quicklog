package main

import (
	"flag"
	"log"

	"github.com/karmakaze/quicklog/web"
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	web.Serve("0.0.0.0")
}
