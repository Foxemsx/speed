package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var (
		themeFlag  = flag.String("theme", "default", "color theme: default")
		compactFlag = flag.Bool("compact", false, "skip the large logo, show tagline only")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of riptide:\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n  riptide\n  riptide --compact\n")
	}
	flag.Parse()

	theme := DefaultTheme
	_ = themeFlag // reserved for future palettes

	m := newAppModel(theme, *compactFlag)
	p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "riptide: %v\n", err)
		os.Exit(1)
	}
}
