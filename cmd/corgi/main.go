package main

import (
	"fmt"
	"os"

	"github.com/mavolin/corgi/cmd/corgi/app"
)

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
