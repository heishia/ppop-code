package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#10B981")
	accentColor    = lipgloss.Color("#F59E0B")
	textColor      = lipgloss.Color("#E5E7EB")
	mutedColor     = lipgloss.Color("#6B7280")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	menuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(textColor)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	chatBubbleUser = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1).
			MarginBottom(1)

	chatBubbleAssistant = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(secondaryColor).
				Padding(0, 1).
				MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)
)
