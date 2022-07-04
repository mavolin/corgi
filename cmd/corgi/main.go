package main

import (
	"log"
	"os"

	"github.com/mavolin/corgi/cmd/corgi/app"
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
