package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "layout-cli",
		Usage: "A CLI tool for various layout operations",
		Commands: []*cli.Command{
			viewCommand,
			analyseCommand,
			optimiseCommand,
			compareCommand,
			rankCommand,
			experimentCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
