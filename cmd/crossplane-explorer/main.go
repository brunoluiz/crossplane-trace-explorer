package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func cmdMain(cmds ...*cli.Command) *cli.Command {
	return &cli.Command{
		Name:     "crossplane-explorer",
		Usage:    "Set of tools to explore your crossplane resources",
		Commands: cmds,
	}
}

func main() {
	if err := cmdMain(
		cmdTrace(),
	).Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
