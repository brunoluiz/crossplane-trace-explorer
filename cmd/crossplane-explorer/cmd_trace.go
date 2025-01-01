package main

import (
	"context"
	"os"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/statusbar"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/table"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/tasker"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func cmdTrace() *cli.Command {
	return &cli.Command{
		Usage: `Explore tracing from Crossplane. Usage is available through arguments or data stream
1. To load it straight from a live resource using the crossplane CLI, do 'crossplane-explorer trace <object name>'
2. To load it from a trace JSON file, do 'crossplane beta trace -o json <> | crossplane-explorer trace --stdin'

Live mode is only available for (1) through the use of --watch / --watch-interval (see flag usage below)`,
		Name:    "trace",
		Aliases: []string{"t"},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "cmd", Usage: "Which binary should it use to generate the JSON trace", Value: "crossplane beta trace -o json"},
			&cli.BoolFlag{Name: "stdin", Aliases: []string{"in"}, Usage: "Specify in case file is piped into stdin"},
			&cli.BoolFlag{Name: "watch", Aliases: []string{"w"}, Usage: "Refresh trace every 10 seconds"},
			&cli.DurationFlag{Name: "watch-interval", Aliases: []string{"wi"}, Usage: "Refresh interval for the watcher feature", Value: 5 * time.Second},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			eg, egCtx := errgroup.WithContext(ctx)
			app := tea.NewProgram(
				explorer.New(
					tree.New(table.New(
						table.WithColumns([]table.Column{
							{Title: explorer.HeaderKeyObject, Width: 60},
							{Title: explorer.HeaderKeyGroup, Width: 30},
							{Title: explorer.HeaderKeySynced, Width: 10},
							{Title: explorer.HeaderKeySyncedLast, Width: 25},
							{Title: explorer.HeaderKeyReady, Width: 10},
							{Title: explorer.HeaderKeyReadyLast, Width: 25},
							{Title: explorer.HeaderKeyMessage, Width: 50},
						}),
						table.WithFocused(true),
						table.WithStyles(func() table.Styles {
							s := table.DefaultStyles()
							s.Selected = lipgloss.NewStyle().
								Foreground(lipgloss.ANSIColor(ansi.Black)).
								Background(lipgloss.ANSIColor(ansi.White))
							return s
						}()),
					)),
					viewer.New(),
					statusbar.New(),
				),
				tea.WithAltScreen(),
				// tea.WithMouseCellMotion(),
				tea.WithContext(egCtx),
			)

			eg.Go(func() error {
				_, err := app.Run()
				return err
			})

			eg.Go(func() error {
				if c.Bool("stdin") {
					app.Send(xplane.NewReaderTraceQuerier(os.Stdin).MustGetTrace())
					return nil
				}

				q := xplane.NewCLITraceQuerier(c.String("cmd"), c.Args().First())
				cb := func() error {
					app.Send(q.MustGetTrace())
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
