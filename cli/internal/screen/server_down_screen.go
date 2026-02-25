package screen

import (
	"github.com/charmbracelet/huh"
)

func NewServerDownForm() *huh.Form {

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("🚫 Server is Down").
				Description(
					"The game server is currently unavailable.\n\n"+
						"🔄 Please try again later.",
				),

			huh.NewConfirm().
				Affirmative("Exit").
				Negative(""),
		),
	).WithShowHelp(false)

	form.NextField()

	return form
}
