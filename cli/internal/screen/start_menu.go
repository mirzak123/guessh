package screen

import (
	"errors"
	"fmt"
	"guessh/internal/game"
	"guessh/internal/protocol"
	"guessh/internal/ui"
	"strconv"
	"strings"

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

func NewStartMenu(matchInfo *game.MatchInfo) (*huh.Form, *bool) {
	confirm := true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[protocol.GameMode]().
				Title("Are you lonely?").
				Options(
					huh.NewOption(GameModeLabels[protocol.SINGLE], protocol.SINGLE),
					huh.NewOption(GameModeLabels[protocol.MULTI_REMOTE], protocol.MULTI_REMOTE),
				).
				Value(&matchInfo.Mode),

			huh.NewSelect[int]().
				Title("How many letters?").
				Options(
					huh.NewOption("Five", 5),
					huh.NewOption("Six", 6),
					huh.NewOption("Seven", 7),
				).
				Value(&matchInfo.WordLen),

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
				Value(&matchInfo.RawTotalRounds),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Player name ").
				Value(&matchInfo.PlayerName).
				Validate(func(str string) error {
					if matchInfo.PlayerName == "" {
						return errors.New("name must not be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Room ID ").
				Value(&matchInfo.RoomID).
				Description("Leave empty to create a new room").
				Validate(func(str string) error {
					matchInfo.RoomID = strings.ToUpper(str)

					if str != "" && len(str) != protocol.ROOM_ID_LENGTH {
						return fmt.Errorf("room ID must be %d characters long", protocol.ROOM_ID_LENGTH)
					}
					return nil
				}).
				CharLimit(protocol.ROOM_ID_LENGTH),
		).WithHideFunc(func() bool {
			return matchInfo.Mode != protocol.MULTI_REMOTE
		}),

		huh.NewGroup(
			huh.NewConfirm().
				TitleFunc(func() string {
					if matchInfo.Mode == protocol.MULTI_REMOTE {
						if matchInfo.RoomID != "" {
							return fmt.Sprintf("Join room %s ?", lipgloss.NewStyle().Foreground(lipgloss.Color(ui.Gray)).Render(matchInfo.RoomID))
						} else {
							return "Create new room?"
						}
					} else {
						return "Start match?"
					}
				}, nil).
				Value(&confirm).
				WithButtonAlignment(lipgloss.Left),

			huh.NewNote().
				Title("Summary").
				DescriptionFunc(func() string {
					labelStyle := lipgloss.NewStyle().Bold(true)
					valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.Gray))

					line := func(label, value string) string {
						return labelStyle.Render(label) + valueStyle.Render(value)
					}

					var lines []string
					lines = append(lines, line("Mode: ", GameModeLabels[matchInfo.Mode]))
					lines = append(lines, line("Word length: ", fmt.Sprintf("%d", matchInfo.WordLen)))
					lines = append(lines, line("Rounds: ", matchInfo.RawTotalRounds))

					if matchInfo.Mode == protocol.MULTI_REMOTE {
						lines = append(lines, line("Player name: ", matchInfo.PlayerName))
						if matchInfo.RoomID != "" {
							lines = append(lines, line("Room ID: ", matchInfo.RoomID))
						}
					}

					return lipgloss.JoinVertical(lipgloss.Left, lines...)

				}, &matchInfo.RoomID).Height(10),
		),
	)

	return form, &confirm
}
