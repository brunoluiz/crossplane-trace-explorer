package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "cmd",
				Usage:       "Which binary should it use to generate the JSON trace",
				DefaultText: "crossplane beta trace -o json",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "Object name to trace",
			},
			&cli.BoolFlag{
				Name:  "stdin",
				Usage: "Specify in case file is piped into stdin",
			},
			&cli.BoolFlag{
				Name:  "watch",
				Usage: "Refresh trace every 10 seconds",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			var r io.Reader

			if c.Bool("stdin") {
				r = os.Stdin
			} else {
				var err error
				r, err = execCrossplane(c.String("cmd"), c.String("name"))
				if err != nil {
					return err
				}
			}

			res, err := xplane.Parse(r)
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

func execCrossplane(cmd string, name string) (io.Reader, error) {
	s := strings.Split(cmd, " ")
	app := s[0]
	args := append(s[1:], name)

	stdout, err := exec.Command(app, args...).Output()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(stdout), nil
}
