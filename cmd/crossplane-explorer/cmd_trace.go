package main

import (
	"context"
	"fmt"
	"os"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v3"
)

func cmdTrace() *cli.Command {
	return &cli.Command{
		Usage:   "Explore tracing from Crossplane. Pipe the `crossplane beta trace -o json <>` output to open the trace view.",
		Name:    "trace",
		Aliases: []string{"t"},
		Flags:   []cli.Flag{},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			res, err := xplane.Parse(os.Stdin)
			if err != nil {
				return fmt.Errorf("Error while parsing Crossplane JSON: %w", err)
			}

			_, err = tea.NewProgram(
				explorer.New(res),
				tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
				tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
			).Run()
			if err != nil {
				return err
			}

			return nil
		},
	}
}
