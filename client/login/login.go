package login

import (
	"log"

	tui "github.com/marcusolsson/tui-go"
)

func GetUserName() string {
	var username string
	user := tui.NewEntry()
	user.SetFocused(true)

	label := tui.NewLabel("Enter your name: ")
	user.SetSizePolicy(tui.Expanding, tui.Maximum)

	userBox := tui.NewHBox(
		label,
		user,
	)
	userBox.SetBorder(true)
	userBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	frame := tui.NewVBox(
		tui.NewSpacer(),
		tui.NewPadder(-4, 0, tui.NewPadder(4, 0, userBox)),
		tui.NewSpacer(),
	)

	ui, err := tui.New(frame)
	if err != nil {
		log.Fatal(err)
	}

	user.OnSubmit(func(e *tui.Entry) {
		if e.Text() != "" {
			username = e.Text()
			e.SetText("")
			ui.Quit()
		}
	})

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}

	return username
}
