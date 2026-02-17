package screen

import (
	"fmt"
	"guessh/internal/ui"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type waitingOpponentModel struct {
	spinner spinner.Model
	roomID  string
}

func NewWaitingOpponentModel(roomID string) *waitingOpponentModel {
	s := spinner.New()
	s.Spinner = ui.Spinners[rand.Intn(len(ui.Spinners))]

	return &waitingOpponentModel{
		spinner: s,
		roomID:  roomID,
	}
}

func (m *waitingOpponentModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *waitingOpponentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	spinnerModel, spinnerCmd := m.spinner.Update(msg)
	m.spinner = spinnerModel
	return m, spinnerCmd
}

func (m *waitingOpponentModel) View() string {
	view := strings.Join([]string{
		fmt.Sprintf("Room ID: %s", func() string {
			if m.roomID == "" {
				return m.spinner.View()
			} else {
				return m.roomID
			}
		}()),
		fmt.Sprintf("Waiting for opponent to join... %s", m.spinner.View()),
	}, "\n")

	return view
}
