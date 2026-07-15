// Package ui provides the terminal user interface for easy-ssh using Bubble Tea.
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#bd93f9")   // Purple
	secondaryColor = lipgloss.Color("#8be9fd") // Cyan
	accentColor    = lipgloss.Color("#ff79c6")   // Pink
	greenColor     = lipgloss.Color("#50fa7b")   // Green
	yellowColor    = lipgloss.Color("#f1fa8c")   // Yellow
	grayColor      = lipgloss.Color("#6272a4")   // Comment/Gray
	whiteColor     = lipgloss.Color("#f8f8f2")   // White

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Background(primaryColor).
			Padding(0, 1).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	warningBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(yellowColor).
			Padding(1, 2).
			Foreground(yellowColor).
			Width(60)

	successBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(greenColor).
			Padding(1, 2).
			Foreground(whiteColor).
			Width(70)

	commandStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(grayColor)

	tickStyle = lipgloss.NewStyle().
			Foreground(greenColor).
			Bold(true)

	crossStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)
)
