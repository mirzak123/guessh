package screen

import (
	"errors"
	"fmt"
	"guessh/internal/protocol"
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
			huh.NewConfirm().
				Title("Start match?").
				DescriptionFunc(func() string {
					return fmt.Sprintf(
						"mode:             \t%s\n"+
							"word length:      \t%d\n"+
							"number of rounds: \t%s",
						GameModeLabels[matchInfo.mode], matchInfo.wordLen, matchInfo.rawTotalRounds)
				}, nil).
				Value(&confirm).WithButtonAlignment(lipgloss.Left),
		).WithHeight(6),
	)

	return form, &confirm
}
