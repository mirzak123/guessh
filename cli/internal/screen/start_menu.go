package screen

import (
	"errors"
	"fmt"
	"guessh/internal/protocol"
	"guessh/internal/ui"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	minRounds = 1
	maxRounds = 5
)

var GameModeLabels = map[protocol.GameMode]string{
	protocol.SINGLE:       "Single player",
	protocol.MULTI_REMOTE: "Two player remote",
}

func NewStartMenu(matchInfo *MatchInfo) (*huh.Form, *bool) {
	confirm := true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[protocol.GameMode]().
				Title("Are you lonely?").
				Options(
					huh.NewOption(GameModeLabels[protocol.SINGLE], protocol.SINGLE),
					huh.NewOption(GameModeLabels[protocol.MULTI_REMOTE], protocol.MULTI_REMOTE),
				).
				Value(&matchInfo.mode),

			huh.NewSelect[int]().
				Title("How many letters?").
				Options(
					huh.NewOption("Five", 5),
					huh.NewOption("Six", 6),
					huh.NewOption("Seven", 7),
				).
				Value(&matchInfo.wordLen),

			huh.NewInput().
				Title(fmt.Sprintf("How many rounds? (%d - %d)", minRounds, maxRounds)).
				Validate(func(str string) error {
					r, err := strconv.Atoi(str)
					if err != nil {
						return errors.New("please enter a valid number")
					}
					if r < minRounds || r > maxRounds {
						return fmt.Errorf("round number must be between %d and %d", minRounds, maxRounds)
					}
					return nil
				}).
				Value(&matchInfo.rawTotalRounds),
		),

		huh.NewGroup(
			huh.NewInput().Title("Player name ").Value(&matchInfo.playerName),
			huh.NewInput().Title("Room ID ").Value(&matchInfo.roomID).Description("Creates a new room if left empty"),
		).WithHideFunc(func() bool {
			return matchInfo.mode != protocol.MULTI_REMOTE
		}),

		huh.NewGroup(
			huh.NewConfirm().
				TitleFunc(func() string {
					if matchInfo.mode == protocol.MULTI_REMOTE {
						if matchInfo.roomID != "" {
							return fmt.Sprintf("Join room %s ?", lipgloss.NewStyle().Foreground(lipgloss.Color(ui.Gray)).Render(matchInfo.roomID))
						} else {
							return "Create new room?"
						}
					} else {
						return "Start match?"
					}
				}, nil).
				Value(&confirm).WithButtonAlignment(lipgloss.Left),

			huh.NewNote().
				Title("Summary").
				DescriptionFunc(func() string {
					labelStyle := lipgloss.NewStyle().Bold(true)
					valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.Gray))

					line := func(label, value string) string {
						return labelStyle.Render(label) + valueStyle.Render(value)
					}

					var lines []string
					lines = append(lines, line("Mode: ", GameModeLabels[matchInfo.mode]))
					lines = append(lines, line("Word length: ", fmt.Sprintf("%d", matchInfo.wordLen)))
					lines = append(lines, line("Rounds: ", matchInfo.rawTotalRounds))

					if matchInfo.mode == protocol.MULTI_REMOTE {
						lines = append(lines, line("Player name: ", matchInfo.playerName))
						lines = append(lines, line("Room ID: ", matchInfo.roomID))
					}

					return lipgloss.JoinVertical(lipgloss.Left, lines...)

				}, &matchInfo.roomID).Height(10),
		),
	)

	return form, &confirm
}
