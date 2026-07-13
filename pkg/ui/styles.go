package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	PrimaryColor = lipgloss.Color("#bd93f9")   // Purple
	SecondaryColor = lipgloss.Color("#8be9fd") // Cyan
	AccentColor  = lipgloss.Color("#ff79c6")   // Pink
	GreenColor   = lipgloss.Color("#50fa7b")   // Green
	YellowColor  = lipgloss.Color("#f1fa8c")   // Yellow
	GrayColor    = lipgloss.Color("#6272a4")   // Comment/Gray
	WhiteColor   = lipgloss.Color("#f8f8f2")   // White

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(WhiteColor).
			Background(PrimaryColor).
			Padding(0, 1).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Italic(true)

	WarningBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(YellowColor).
			Padding(1, 2).
			Foreground(YellowColor).
			Width(60)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(GreenColor).
			Padding(1, 2).
			Foreground(WhiteColor).
			Width(70)

	CommandStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	StatusStyle = lipgloss.NewStyle().
			Foreground(GrayColor)

	TickStyle = lipgloss.NewStyle().
			Foreground(GreenColor).
			Bold(true)

	CrossStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)
)
