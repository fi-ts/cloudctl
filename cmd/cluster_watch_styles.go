package cmd

import "charm.land/lipgloss/v2"

// lipgloss styles used by cluster watch
var (
	styleTopBox = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff4500")).
		Align(lipgloss.Center).
		Bold(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff4500"))
)
