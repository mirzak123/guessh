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
	minRounds      = 1
	maxRounds      = 5
	minTurnTimeout = 5
	maxTurnTimeout = 300
)

var GameModeLabels = map[protocol.GameMode]string{
	protocol.SINGLE:       "Single Player",
	protocol.MULTI_LOCAL:  "Local Multiplayer",
	protocol.MULTI_REMOTE: "Remote Multiplayer",
}

var GameModeDescriptions = map[protocol.GameMode]string{
	protocol.SINGLE:       "Yes...",
	protocol.MULTI_LOCAL:  "No, we're sharing a chair.",
	protocol.MULTI_REMOTE: "No, they're in the cloud.",
}

var GameFormatLabels = map[protocol.GameFormat]string{
	protocol.WORDLE:  "Wordle",
	protocol.QUORDLE: "Quordle",
}

func NewGameConfigMenu(matchInfo *game.MatchInfo, hoveredPtr *protocol.GameMode) (*huh.Form, *bool) {
	confirm := true

	var (
		playerNameInput = huh.NewInput().
				TitleFunc(func() string {
				if matchInfo.Mode == protocol.MULTI_LOCAL {
					return "Player 1"
				} else {
					return "Player name"
				}
			}, matchInfo.Mode).
			Value(&matchInfo.PlayerName).
			Validate(func(str string) error {
				if matchInfo.PlayerName == "" {
					return errors.New("name must not be empty")
				}
				return nil
			})

		opponentNameInput = huh.NewInput().
					Title("Player 2").
					Value(&matchInfo.OpponentName).
					Validate(func(str string) error {
				if matchInfo.OpponentName == "" {
					return errors.New("player name must not be empty")
				}
				return nil
			})

		modeInput = huh.NewSelect[protocol.GameMode]().
				Title("Are you lonely?").
				Options(
				huh.NewOption(GameModeLabels[protocol.SINGLE], protocol.SINGLE),
				huh.NewOption(GameModeLabels[protocol.MULTI_LOCAL], protocol.MULTI_LOCAL),
				huh.NewOption(GameModeLabels[protocol.MULTI_REMOTE], protocol.MULTI_REMOTE),
			).
			DescriptionFunc(func() string {
				return GameModeDescriptions[*hoveredPtr]
			}, hoveredPtr).
			Value(&matchInfo.Mode)

		formatInput = huh.NewSelect[protocol.GameFormat]().
				Title("Game format?").
				Options(
				huh.NewOption(GameFormatLabels[protocol.WORDLE], protocol.WORDLE),
				huh.NewOption(GameFormatLabels[protocol.QUORDLE], protocol.QUORDLE),
			).
			Value(&matchInfo.Format)

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

		timerInput = huh.NewInput().
				Title(fmt.Sprintf("How many seconds per turn? (%d - %d)", minTurnTimeout, maxTurnTimeout)).
				Description("Leave empty for no time limit").
				Validate(func(str string) error {
				if str == "" {
					return nil
				}

				t, err := strconv.Atoi(str)
				if err != nil {
					return errors.New("please enter a valid number")
				}
				if t < minTurnTimeout || t > maxTurnTimeout {
					return fmt.Errorf("seconds per turn must be between %d and %d", minTurnTimeout, maxTurnTimeout)
				}
				return nil
			}).
			Value(&matchInfo.RawTurnTimeout)

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

				switch matchInfo.Mode {
				case protocol.SINGLE:
					lines = append(lines, line("Game format: ", GameFormatLabels[matchInfo.Format]))
					lines = append(lines, line("Word length: ", fmt.Sprintf("%d", matchInfo.WordLen)))
					lines = append(lines, line("Rounds: ", matchInfo.RawTotalRounds))

					if matchInfo.RawTurnTimeout != "" {
						lines = append(lines, line("Seconds per turn: ", matchInfo.RawTurnTimeout))
					}

				case protocol.MULTI_LOCAL:
					lines = append(lines, line("Player 1: ", matchInfo.PlayerName))
					lines = append(lines, line("Player 2: ", matchInfo.OpponentName))
					lines = append(lines, line("Game format: ", GameFormatLabels[matchInfo.Format]))
					lines = append(lines, line("Word length: ", fmt.Sprintf("%d", matchInfo.WordLen)))
					lines = append(lines, line("Rounds: ", matchInfo.RawTotalRounds))

					if matchInfo.RawTurnTimeout != "" {
						lines = append(lines, line("Seconds per turn: ", matchInfo.RawTurnTimeout))
					}

				case protocol.MULTI_REMOTE:
					lines = append(lines, line("Player name: ", matchInfo.PlayerName))
					if matchInfo.JoinExisting {
						lines = append(lines, line("Room ID: ", matchInfo.RoomID))
					} else {
						lines = append(lines, line("Game format: ", GameFormatLabels[matchInfo.Format]))
						lines = append(lines, line("Word length: ", fmt.Sprintf("%d", matchInfo.WordLen)))
						lines = append(lines, line("Rounds: ", matchInfo.RawTotalRounds))

						if matchInfo.RawTurnTimeout != "" {
							lines = append(lines, line("Seconds per turn: ", matchInfo.RawTurnTimeout))
						}
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
			formatInput,
			wordLenInput,
			roundNumInput,
			timerInput,
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
			playerNameInput,
			opponentNameInput,
		).WithHideFunc(func() bool {
			return matchInfo.Mode != protocol.MULTI_LOCAL
		}),

		huh.NewGroup(
			cornfirmInput,
			summaryNote,
		),
	)

	return form, &confirm
}
