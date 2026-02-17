package ui

import "github.com/charmbracelet/lipgloss"

var (
	Green  = lipgloss.Color("46")
	Yellow = lipgloss.Color("226")
	Gray   = lipgloss.Color("7")
	White  = lipgloss.Color("15")
	Purple = lipgloss.Color("63")
)

var (
	MainContentStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Purple).
				Padding(1, 2)

	baseLetterStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Bold(true).
			Padding(0, 1)

	correctStyle = baseLetterStyle.
			BorderForeground(Green).
			Foreground(Green)

	presentStyle = baseLetterStyle.
			BorderForeground(Yellow).
			Foreground(Yellow)

	absentStyle = baseLetterStyle.
			BorderForeground(Gray).
			Foreground(Gray)

	activeInputStyle = baseLetterStyle.
				BorderForeground(Purple).
				Foreground(White)
)
