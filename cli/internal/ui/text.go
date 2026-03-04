package ui

import "github.com/charmbracelet/lipgloss"

func ASCIILogo() string {
	gue := `
 ____ ____ ____ 
||G |||u |||e ||
||__|||__|||__||
|/__\|/__\|/__\|`

	ssh := `
 ____ ____ ____ 
||S |||S |||H ||
||__|||__|||__||
|/__\|/__\|/__\|`

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		gue,
		lipgloss.NewStyle().Foreground(Purple).Render(ssh),
	)
}
