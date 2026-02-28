package screen

import (
	"fmt"
	"guessh/internal/game"
	"guessh/internal/ui"
	"math/rand"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type waitingOpponentModel struct {
	spinner spinner.Model
	form    *huh.Form
	roomID  string
}

func NewWaitingOpponentModel() *waitingOpponentModel {
	s := spinner.New()
	s.Spinner = ui.Spinners[rand.Intn(len(ui.Spinners))]

	m := &waitingOpponentModel{
		spinner: s,
	}

	m.form = m.buildForm()

	return m
}

func (m *waitingOpponentModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.form.Init(),
	)
}

func (m *waitingOpponentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	cmds = append(cmds, spinnerCmd)

	m.form = m.buildForm()

	if form, formCmd := m.form.Update(msg); formCmd != nil {
		m.form = form.(*huh.Form)
		cmds = append(cmds, formCmd)
	}

	if m.form.State == huh.StateCompleted {
		return m, emit(game.LeaveMatchIntent{})
	}

	return m, tea.Batch(cmds...)
}

func (m *waitingOpponentModel) View() string {
	return m.form.View()
}

func (m *waitingOpponentModel) buildForm() *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("🏠 Room").
				Description(func() string {
					var roomID string
					if m.roomID == "" {
						roomID = m.spinner.View()
					} else {
						roomID = lipgloss.NewStyle().
							Border(lipgloss.RoundedBorder()).
							BorderForeground(ui.Purple).
							Padding(0, 1).
							Bold(true).
							Render(m.roomID)
					}
					return lipgloss.JoinHorizontal(lipgloss.Center, "Room ID: ", roomID)
				}()),

			huh.NewNote().
				Title("Status").
				Description(
					fmt.Sprintf(
						"Waiting for opponent to join... %s",
						m.spinner.View(),
					),
				),

			huh.NewConfirm().
				Title("").
				Affirmative("Leave room").
				Negative(""),
		),
	)

	form.NextField()

	return form
}
