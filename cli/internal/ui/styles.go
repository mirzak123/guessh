package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	Green  = lipgloss.Color("46")
	Yellow = lipgloss.Color("226")
	Gray   = lipgloss.Color("244")
	White  = lipgloss.Color("15")
	Purple = lipgloss.Color("63")
	Rose   = lipgloss.Color("198")
)

var (
	Theme = huh.ThemeCharm()

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

	playerActiveInputStyle = baseLetterStyle.
				BorderForeground(Purple).
				Foreground(White)

	opponentActiveInputStyle = baseLetterStyle.
					BorderForeground(Rose).
					Foreground(White)

	PurpleText = lipgloss.NewStyle().
			Foreground(Purple)

	RoseText = lipgloss.NewStyle().
			Foreground(Rose)

	YellowText = lipgloss.NewStyle().
			Foreground(Yellow)

	GreenText = lipgloss.NewStyle().
			Foreground(Green)

	GrayText = lipgloss.NewStyle().
			Foreground(Gray)

	WhiteText = lipgloss.NewStyle().
			Foreground(White)
)

var (
	symbolCheck = "✓"
	symbolCross = "✗"
	symbolDraw  = "○"
	symbolNone  = "∅"
)
