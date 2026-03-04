package screen

import (
	"errors"
	"fmt"
	"guessh/internal/game"
	"guessh/internal/logger"
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

func NewGameConfigMenu(matchInfo *game.MatchInfo) (*huh.Form, *bool) {
	confirm := true

	var (
		playerNameInput = huh.NewInput().
				Title("Player name ").
				Value(&matchInfo.PlayerName).
				Validate(func(str string) error {
				if matchInfo.PlayerName == "" {
					return errors.New("name must not be empty")
				}
				return nil
			})

		modeInput = huh.NewSelect[protocol.GameMode]().
				Title("Are you lonely?").
				Options(
				huh.NewOption(GameModeLabels[protocol.SINGLE], protocol.SINGLE),
				huh.NewOption(GameModeLabels[protocol.MULTI_REMOTE], protocol.MULTI_REMOTE),
			).
			Value(&matchInfo.Mode)

		joinExistingInput = huh.NewSelect[bool]().
					Title("Create or join?").
					Options(
				huh.NewOption("Create new match", false),
				huh.NewOption("Join existing match", true),
			).
			Value(&matchInfo.JoinExisting)

		wordLenInput = huh.NewSelect[int]().
				Title("How many letters?").
				Options(
				huh.NewOption("Five", 5),
				huh.NewOption("Six", 6),
				huh.NewOption("Seven", 7),
			).
			Value(&matchInfo.WordLen)

		roundNumInput = huh.NewInput().
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
			Value(&matchInfo.RawTotalRounds)

		roomIDInput = huh.NewInput().
				Title("Room ID ").
				Value(&matchInfo.RoomID).
				Description("Enter the Room ID shared with you").
				Validate(func(str string) error {
				if matchInfo.RoomValidationError != nil {
					err := matchInfo.RoomValidationError
					matchInfo.RoomValidationError = nil
					logger.Debug("Error: %v", err)
					return err
				}

				matchInfo.RoomID = strings.ToUpper(str)
				if str != "" && len(str) != protocol.ROOM_ID_LENGTH {
					return fmt.Errorf("room ID must be %d characters long", protocol.ROOM_ID_LENGTH)
				}
				return nil
			}).
			CharLimit(protocol.ROOM_ID_LENGTH)

		cornfirmInput = huh.NewConfirm().
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
			WithButtonAlignment(lipgloss.Left)

		summaryNote = huh.NewNote().
				Title("Summary").
				DescriptionFunc(func() string {
				labelStyle := lipgloss.NewStyle().Bold(true)
				valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.Gray))

				line := func(label, value string) string {
					return labelStyle.Render(label) + valueStyle.Render(value)
				}

				var lines []string
				lines = append(lines, line("Mode: ", GameModeLabels[matchInfo.Mode]))

				if !matchInfo.JoinExisting {
					lines = append(lines, line("Word length: ", fmt.Sprintf("%d", matchInfo.WordLen)))
					lines = append(lines, line("Rounds: ", matchInfo.RawTotalRounds))
				}

				if matchInfo.Mode == protocol.MULTI_REMOTE {
					lines = append(lines, line("Player name: ", matchInfo.PlayerName))
					if matchInfo.RoomID != "" {
						lines = append(lines, line("Room ID: ", matchInfo.RoomID))
					}
				}

				return lipgloss.JoinVertical(lipgloss.Left, lines...)

			}, &matchInfo).Height(10)
	)

	form := huh.NewForm(
		huh.NewGroup(modeInput).
			WithShowHelp(false),

		huh.NewGroup(joinExistingInput).
			WithHideFunc(func() bool {
				return matchInfo.Mode != protocol.MULTI_REMOTE
			}),

		huh.NewGroup(
			wordLenInput,
			roundNumInput,
		).WithHideFunc(func() bool {
			return matchInfo.JoinExisting
		}),

		huh.NewGroup(
			playerNameInput,
			roomIDInput,
		).WithHideFunc(func() bool {
			return matchInfo.Mode != protocol.MULTI_REMOTE || !matchInfo.JoinExisting
		}),

		huh.NewGroup(
			playerNameInput,
		).WithHideFunc(func() bool {
			return matchInfo.Mode != protocol.MULTI_REMOTE || matchInfo.JoinExisting
		}),

		huh.NewGroup(
			cornfirmInput,
			summaryNote,
		),
	)

	return form, &confirm
}
