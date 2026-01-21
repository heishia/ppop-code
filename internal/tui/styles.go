package tui

import (
	"os"
	"runtime"

	"github.com/charmbracelet/lipgloss"
)

// useASCIIBorder determines if we should use ASCII borders for compatibility
// Set PPOPCODE_ASCII=1 to force ASCII mode, or it auto-detects on Windows
func useASCIIBorder() bool {
	// Allow override via environment variable
	if os.Getenv("PPOPCODE_ASCII") == "1" {
		return true
	}
	if os.Getenv("PPOPCODE_ASCII") == "0" {
		return false
	}
	// Check for Windows Terminal or modern terminals that support Unicode
	if os.Getenv("WT_SESSION") != "" || os.Getenv("TERM_PROGRAM") != "" {
		return false
	}
	// Default to ASCII on Windows (legacy cmd/PowerShell)
	return runtime.GOOS == "windows"
}

// getBorder returns appropriate border style based on terminal compatibility
func getBorder() lipgloss.Border {
	if useASCIIBorder() {
		return lipgloss.NormalBorder()
	}
	return lipgloss.RoundedBorder()
}

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
			Border(getBorder()).
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
			Border(getBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1).
			MarginBottom(1)

	chatBubbleAssistant = lipgloss.NewStyle().
				Border(getBorder()).
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
