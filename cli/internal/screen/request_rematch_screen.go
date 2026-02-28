package screen

import (
	"guessh/internal/game"
	"guessh/internal/ui"
	"math/rand"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type requestRematchModel struct {
	spinner               spinner.Model
	opponentName          string
	opponentDeniedRematch bool
	form                  *huh.Form
}

func NewRequestRematchModel(opponentName string) *requestRematchModel {
	s := spinner.New()
	s.Spinner = ui.Spinners[rand.Intn(len(ui.Spinners))]

	m := &requestRematchModel{
		spinner:      s,
		opponentName: opponentName,
	}

	m.form = m.buildForm()

	return m
}

func (m *requestRematchModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.form.Init(),
	)
}

func (m *requestRematchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.opponentDeniedRematch {
			return m, emit(game.StartGameIntent{})
		} else {
			return m, emit(game.DenyRematchIntent{})
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *requestRematchModel) View() string {
	return m.form.View()
}

func (m *requestRematchModel) buildForm() *huh.Form {
	var (
		title       string
		description string
		confirmText string
		opponent    = lipgloss.NewStyle().Foreground(ui.Rose).Render(m.opponentName)
	)

	if m.opponentDeniedRematch {
		title = lipgloss.JoinHorizontal(lipgloss.Center, "🚫 ", opponent, " denied the rematch")
		description = "Press enter to continue to main screen"
		confirmText = "Continue"
	} else {
		title = "🤔 Rematch Requested"
		description = lipgloss.JoinHorizontal(
			lipgloss.Center,
			"Waiting for ",
			opponent,
			" to confirm rematch... ",
			m.spinner.View(),
		)
		confirmText = "Abort rematch"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(title).
				Description(description),

			huh.NewConfirm().
				Title("").
				Affirmative(confirmText).
				Negative(""),
		),
	).WithShowHelp(false)

	form.NextField()

	return form
}
