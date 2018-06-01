package main

import (
	"fmt"
	"os"

	"github.com/DrWrong/monica"
	cli "gopkg.in/urfave/cli.v2"
)

func main() {
	app := monica.App
	app.Action = func(c *cli.Context) error {
		fmt.Println("Hello word")
		return nil
	}
	app.Run(os.Args)
}
