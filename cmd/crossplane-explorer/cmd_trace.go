package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer"
	"github.com/brunoluiz/crossplane-explorer/internal/tasker"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func cmdTrace() *cli.Command {
	return &cli.Command{
		Usage:   "Explore tracing from Crossplane. Pipe the `crossplane beta trace -o json <>` output to open the trace view.",
		Name:    "trace",
		Aliases: []string{"t"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "cmd",
				Usage: "Which binary should it use to generate the JSON trace",
				Value: "crossplane beta trace -o json",
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
			&cli.DurationFlag{
				Name:  "watch-interval",
				Usage: "Refresh interval for the watcher feature",
				Value: 5 * time.Second,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
			defer stop()

			eg, egCtx := errgroup.WithContext(ctx)
			app := tea.NewProgram(
				explorer.New(),
				tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
				tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
				tea.WithContext(egCtx),
			)

			eg.Go(func() error {
				_, err := app.Run()
				stop()
				return err
			})

			eg.Go(func() error {
				cb := func() error {
					r, err := execCrossplane(c.String("cmd"), c.String("name"))
					if err != nil {
						return err
					}

					app.Send(xplane.MustParse(r))
					return nil
				}

				if c.Bool("stdin") {
					app.Send(xplane.MustParse(os.Stdin))
					return nil
				}

				if !c.Bool("watch") {
					return cb()
				}

				return tasker.Periodic(egCtx, c.Duration("watch-interval"), cb)
			})

			return eg.Wait()
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
