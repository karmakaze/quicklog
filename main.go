package main

import (
	"fmt"
	"github.com/karmakaze/quicklog/web"
)

func main() {
    dbUrl := "user=postgres password=postgres host=127.0.0.1 dbname=quicklog sslmode=disable"
	if err := web.Serve(8080, dbUrl); err != nil {
		fmt.Println(err.Error())
	}
}
