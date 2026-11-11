package cmd

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

func rowStyle(row, col int) lipgloss.Style {
	switch {
	case row == table.HeaderRow:
		return headerStyle
	default:
		return lipgloss.NewStyle()
	}
}

// lipgloss styles used by cluster watch
var (
	topRow = table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().
			Foreground(
				lipgloss.Color("#ffffff"))).
		StyleFunc(rowStyle)

	conditionRow = table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().
				BorderForeground(
				lipgloss.Color("#696969"))).
		StyleFunc(rowStyle)

	nodeRow = table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().
			BorderForeground(
				lipgloss.Color("#ffffff"))).
		StyleFunc(rowStyle)

	errorRow = table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().
				BorderForeground(
				lipgloss.Color("#ff4500"))).
		StyleFunc(rowStyle)

	headerStyle = lipgloss.NewStyle().Bold(true)

	canvas = lipgloss.NewStyle()

	marginTop = lipgloss.NewStyle().MarginTop(1)

	bold = lipgloss.NewStyle().Bold(true)

	red = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff0000"))

	darkorange = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff4500"))

	orange = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff8c00"))

	gold = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffd700"))

	yellow = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffff00"))

	green = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7fff00"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#242424"))

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Background(lipgloss.Color("#242424")).
			PaddingLeft(4)

	keyCanvas = lipgloss.NewStyle().
			Background(lipgloss.Color("#242424"))
)
