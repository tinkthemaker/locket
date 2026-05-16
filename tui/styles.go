package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	cPrimary = lipgloss.Color("#f59e0b")
	cAccent  = lipgloss.Color("#38bdf8")
	cSuccess = lipgloss.Color("#22c55e")
	cDanger  = lipgloss.Color("#ef4444")
	cMuted   = lipgloss.Color("#9ca3af")
	cSubtle  = lipgloss.Color("#52525b")

	titleStyle = lipgloss.NewStyle().
			Foreground(cPrimary).
			Bold(true)

	taglineStyle = lipgloss.NewStyle().
			Foreground(cMuted).
			Italic(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cSubtle).
			Padding(1, 4)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cSubtle).
			Padding(1, 3)

	headingStyle = lipgloss.NewStyle().
			Foreground(cPrimary).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(cMuted).
			Bold(true)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(cAccent).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(cMuted)

	errorStyle = lipgloss.NewStyle().
			Foreground(cDanger).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(cSuccess).
			Bold(true)

	revealValueStyle = lipgloss.NewStyle().
				Foreground(cPrimary).
				Bold(true)

	revealLabelStyle = lipgloss.NewStyle().
				Foreground(cMuted)

	dangerLabelStyle = lipgloss.NewStyle().
				Foreground(cDanger).
				Bold(true)
)

func banner() string {
	title := titleStyle.Render("◆  l o c k e t")
	tagline := taglineStyle.Render("tiny encrypted vault")
	inner := lipgloss.JoinVertical(lipgloss.Center, title, tagline)
	return boxStyle.Render(inner)
}

func smallBanner() string {
	return titleStyle.Render("◆  locket")
}

func helpBar(pairs ...[2]string) string {
	var parts []string
	for _, p := range pairs {
		parts = append(parts, keyHintStyle.Render(p[0])+" "+helpStyle.Render(p[1]))
	}
	return helpStyle.Render(strings.Join(parts, "   ·   "))
}

func center(width, height int, content string) string {
	if width <= 0 || height <= 0 {
		return content
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
