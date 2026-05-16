package tui

import "github.com/charmbracelet/lipgloss"

var (
	revealStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Padding(0, 2)

	revealLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Padding(0, 2)
)
