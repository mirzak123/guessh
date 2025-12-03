package screen

import (
	"errors"
	"fmt"
	"guessh/cmd/cli/protocol"
	"strconv"

	"github.com/charmbracelet/huh"
)

const (
	minRounds = 1
	maxRounds = 5
)

var (
	RawRounds string
	Mode      protocol.GameMode
	WordLen   int
	Start     bool
)

var StartMenu = huh.NewForm(
	huh.NewGroup(
		huh.NewSelect[protocol.GameMode]().
			Title("Are you lonely?").
			Options(
				huh.NewOption("Single player", protocol.SINGLE),
				huh.NewOption("Two player local", protocol.MULTI_LOCAL),
				huh.NewOption("Two player remote", protocol.MULTI_REMOTE),
			).
			Value(&Mode),

		huh.NewSelect[int]().
			Title("How big of words are we feeling?").
			Options(
				huh.NewOption("Five", 5),
				huh.NewOption("Six", 6),
				huh.NewOption("Seven", 7),
			).
			Value(&WordLen),

		huh.NewInput().
			Title(fmt.Sprintf("How many rounds? (%d - %d)", minRounds, maxRounds)).
			Prompt("-> ").
			Validate(func(str string) error {
				r, err := strconv.Atoi(str)
				if err != nil {
					return errors.New("Please enter a valid number")
				}
				if r < minRounds || r > maxRounds {
					return fmt.Errorf("Round number must be between %d and %d", minRounds, maxRounds)
				}
				return nil
			}).
			Value(&RawRounds),
	),

	huh.NewGroup(
		huh.NewConfirm().
			Title("Start match?").
			Value(&Start),
	),
)
