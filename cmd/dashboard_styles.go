package cmd

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

// dashColor is the color type used by the dashboard render helpers.
type dashColor = color.Color

// colors and styles used by the dashboard
var (
	dashBorderColor = lipgloss.Color("#696969")
	dashBarEmpty    = lipgloss.Color("#3a3a3a")

	dashCloseColor = lipgloss.Color("#7fff00")
	dashWarnColor  = lipgloss.Color("#ffff00")
	dashBadColor   = lipgloss.Color("#ff0000")

	dashActiveTab   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffff00"))
	dashInactiveTab = lipgloss.NewStyle().Foreground(lipgloss.Color("#a9a9a9"))
	dashSelected    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))

	dashContentStyle = lipgloss.NewStyle()

	dashPalette = []dashColor{
		lipgloss.Color("#7fff00"),
		lipgloss.Color("#ffff00"),
		lipgloss.Color("#ff8c00"),
		lipgloss.Color("#ff4500"),
		lipgloss.Color("#ff0000"),
		lipgloss.Color("#00bfff"),
		lipgloss.Color("#da70d6"),
		lipgloss.Color("#00ff7f"),
	}
)

// dashboardApplyTheme adapts the dashboard styles to the requested theme.
func dashboardApplyTheme(theme string) error {
	switch theme {
	case "default":
		dashContentStyle = lipgloss.NewStyle()
		dashActiveTab = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffff00"))
		dashInactiveTab = lipgloss.NewStyle().Foreground(lipgloss.Color("#a9a9a9"))
	case "dark":
		dashContentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
		dashActiveTab = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b8860b"))
		dashInactiveTab = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	default:
		return fmt.Errorf("unknown theme: %s", theme)
	}
	return nil
}

// dashPercentColor returns a color for a health-ish percentage
func dashPercentColor(percent int) dashColor {
	switch {
	case percent >= 90:
		return dashCloseColor
	case percent >= 70:
		return dashWarnColor
	default:
		return dashBadColor
	}
}

// dashPanel wraps content in a rounded border and renders a title caption on
// top of the border, mirroring the termui block look of the old dashboard.
func dashPanel(title, content string, width int) string {
	if width < 4 {
		width = 4
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dashBorderColor).
		Width(width-2).
		Padding(0, 1).
		Render(dashContentStyle.Render(content))

	if title == "" {
		return box
	}

	return addCaption(box, title)
}

// dashBarChart renders a simple horizontal bar chart.
func dashBarChart(labels []string, values []float64, colors []dashColor, width int) string {
	if len(labels) == 0 {
		return ""
	}

	maxValue := 0.0
	labelWidth := 0
	for i, label := range labels {
		if len(label) > labelWidth {
			labelWidth = len(label)
		}
		if i < len(values) && values[i] > maxValue {
			maxValue = values[i]
		}
	}

	barWidth := width - labelWidth - 10
	if barWidth < 4 {
		barWidth = 4
	}

	var b strings.Builder
	for i, label := range labels {
		value := 0.0
		if i < len(values) {
			value = values[i]
		}

		filled := 0
		if maxValue > 0 {
			filled = int(math.Round(value / maxValue * float64(barWidth)))
		}

		var bar strings.Builder
		if filled > 0 {
			color := colors[i%len(colors)]
			bar.WriteString(lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)))
		}
		if filled < barWidth {
			bar.WriteString(lipgloss.NewStyle().Foreground(dashBarEmpty).Render(strings.Repeat("░", barWidth-filled)))
		}

		fmt.Fprintf(&b, "%*s %s %s\n", labelWidth, label, bar.String(), dashNumber(value))
	}

	return strings.TrimRight(b.String(), "\n")
}

// dashGauge renders a labeled percentage bar.
func dashGauge(label string, percent int, color dashColor, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	const labelWidth = 12

	barWidth := width - labelWidth - 7
	if barWidth < 4 {
		barWidth = 4
	}

	filled := percent * barWidth / 100
	bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(dashBarEmpty).Render(strings.Repeat("░", barWidth-filled))

	return fmt.Sprintf("%-*s %s %3d%%", labelWidth, label, bar, percent)
}

// dashNumber formats a count without a trailing decimal.
func dashNumber(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.1f", v)
}

// pieEntry is a single line of a rendered pie/legend chart.
type pieEntry struct {
	Label   string
	Count   int
	Percent float64
}

// dashPie renders a pie chart as a sorted, colored legend with mini bars.
func dashPie(entries []pieEntry, width int) string {
	if len(entries) == 0 {
		return ""
	}

	labelWidth := 0
	for _, e := range entries {
		if len(e.Label) > labelWidth {
			labelWidth = len(e.Label)
		}
	}

	barWidth := width - labelWidth - 10
	if barWidth < 4 {
		barWidth = 4
	}

	var b strings.Builder
	for i, e := range entries {
		filled := int(math.Round(e.Percent * float64(barWidth)))
		bar := lipgloss.NewStyle().Foreground(dashPalette[i%len(dashPalette)]).Render(strings.Repeat("█", filled))

		fmt.Fprintf(&b, "%*s %s %3d%%\n", labelWidth, e.Label, bar, int(math.Round(e.Percent*100)))
	}

	return strings.TrimRight(b.String(), "\n")
}

// dashMaxLines truncates s to at most n lines.
func dashMaxLines(s string, n int) string {
	if n <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}

// dashTable renders a bordered table with optional headers. When height is
// greater than zero, the table is limited to that many rows.
func dashTable(headers []string, rows [][]string, width, height int) string {
	if width < 8 {
		width = 8
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().BorderForeground(dashBorderColor)).
		BaseStyle(dashContentStyle).
		StyleFunc(rowStyle).
		Width(width)

	if len(headers) > 0 {
		t = t.Headers(headers...)
	}
	t = t.Rows(rows...)

	if height > 0 {
		t = t.Height(height)
	}

	return t.Render()
}
