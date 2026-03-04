package screen

import (
	"fmt"
	"guessh/internal/game"
	"guessh/internal/ui"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type startMenuModel struct {
	selected int
	options  []string
	width    int
	height   int
}

func NewStartMenu() *startMenuModel {
	return &startMenuModel{
		selected: 0,
		options:  []string{"Play", "Exit"},
	}
}

func (m *startMenuModel) Init() tea.Cmd {
	return nil
}

func (m *startMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "left", "h":
			if m.selected > 0 {
				m.selected--
			} else {
				m.selected = len(m.options) - 1
			}

		case "right", "l":
			m.selected = (m.selected + 1) % len(m.options)

		case "enter":
			switch m.selected {
			case 0:
				return m, emit(game.PlayGameIntent{})
			case 1:
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m *startMenuModel) View() string {
	repoUrl := "https://github.com/mirzak123/guessh/"

	welcomeMessage := fmt.Sprintf(
		"Welcome to %s!\n",
		ui.SmallLogo(),
	)

	shamelessPlug := lipgloss.JoinVertical(
		lipgloss.Center,
		"Enjoying the game? Give it a ⭐ on GitHub:",
		hyperlink(repoUrl, repoUrl),
	)

	var buttons []string

	for i, option := range m.options {
		if m.selected == i {
			buttons = append(buttons, ui.Theme.Focused.FocusedButton.Render(option))
		} else {
			buttons = append(buttons, ui.Theme.Blurred.BlurredButton.Render(option))
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		ui.ASCIILogo(),
		"\n",
		welcomeMessage,
		lipgloss.NewStyle().Foreground(ui.Gray).Italic(true).Render(shamelessPlug),
		"\n",
		lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(buttons, " ")),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func hyperlink(url, text string) string {
	return "\x1b]8;;" + url + "\x1b\\" + text +
		"\x1b]8;;\x1b\\"
}
