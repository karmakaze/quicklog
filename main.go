package main

import (
	"fmt"

	"github.com/karmakaze/quicklog/config"
	"github.com/karmakaze/quicklog/web"
)

func main() {
	cfg := config.Parse()

	err := web.Serve(cfg)
	if err != nil {
		fmt.Println(err.Error())
	}
}
