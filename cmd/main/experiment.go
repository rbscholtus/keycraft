package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var experimentCommand = &cli.Command{
	Name:   "experiment",
	Usage:  "Run experiments",
	Action: experimentAction,
}

func experimentAction(c *cli.Context) error {
	fmt.Println("Running experiment...")
	return nil
}
