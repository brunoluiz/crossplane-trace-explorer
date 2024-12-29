package main

import (
	"fmt"
	"os"

	"github.com/brunoluiz/crossplane-trace-explorer/internal/bubbles/explorer"
	"github.com/brunoluiz/crossplane-trace-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	res, err := xplane.Parse(os.Stdin)
	if err != nil {
		fmt.Printf("Error while parsing Crossplane JSON: %s\n", err)
		os.Exit(1)
	}

	_, err = tea.NewProgram(
		explorer.New(res),
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	).Run()
	if err != nil {
		os.Exit(1)
	}
}
