package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#A78BFA")
	accentColor    = lipgloss.Color("#10B981")
	errorColor     = lipgloss.Color("#EF4444")
	mutedColor     = lipgloss.Color("#6B7280")
	bgColor        = lipgloss.Color("#1F2937")

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Subtitle/help text
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginBottom(1)

	// Input field style
	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	// Selected item in list
	SelectedStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Normal item in list
	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	// Dimmed/description text
	DimStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Loading/spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Help bar at bottom
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Status message style
	StatusStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// Progress bar styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(accentColor)

	// Box for content
	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2)
)
