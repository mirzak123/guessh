package screen

import (
	"guessh/internal/game"
	"guessh/internal/ui"

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

		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}

		case "down", "j":
			if m.selected < len(m.options)-1 {
				m.selected++
			}

		case "enter":
			switch m.selected {
			case 0:
				return m, emit(game.StartGameIntent{})
			case 1:
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m *startMenuModel) View() string {
	buttons := lipgloss.JoinVertical(lipgloss.Center, m.options...)

	content := ui.ASCIILogo() + "\n\n" + buttons

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
