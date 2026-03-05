package graph

import "github.com/charmbracelet/lipgloss"

// Grafana color palette for chart series.
//
//nolint:gochecknoglobals
var grafanaColors = []lipgloss.Color{
	lipgloss.Color("#7EB26D"), // Green
	lipgloss.Color("#EAB839"), // Yellow
	lipgloss.Color("#6ED0E0"), // Cyan
	lipgloss.Color("#EF843C"), // Orange
	lipgloss.Color("#E24D42"), // Red
	lipgloss.Color("#1F78C1"), // Blue
	lipgloss.Color("#BA43A9"), // Purple
	lipgloss.Color("#705DA0"), // Violet
	lipgloss.Color("#508642"), // Dark Green
	lipgloss.Color("#CCA300"), // Gold
}

// ColorForIndex returns the color for a given series index.
func ColorForIndex(idx int) lipgloss.Color {
	return grafanaColors[idx%len(grafanaColors)]
}

// Compliance status colors.
//
//nolint:gochecknoglobals
var (
	colorComplianceOK      = lipgloss.Color("#73BF69") // Green — meeting target
	colorComplianceWarning = lipgloss.Color("#FADE2A") // Yellow — just below target
	colorComplianceDanger  = lipgloss.Color("#FF9830") // Orange — moderately below
	colorComplianceCrit    = lipgloss.Color("#F2495C") // Red — significantly breaching
)

// ComplianceColor returns a color reflecting how close value is to target.
// Both value and target are percentages (0–100).
func ComplianceColor(value, target float64) lipgloss.Color {
	if target <= 0 {
		target = 100 // No target: grade against 100%
	}
	ratio := value / target
	switch {
	case ratio >= 1.0:
		return colorComplianceOK
	case ratio >= 0.99:
		return colorComplianceWarning
	case ratio >= 0.95:
		return colorComplianceDanger
	default:
		return colorComplianceCrit
	}
}
