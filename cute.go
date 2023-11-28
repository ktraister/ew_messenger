package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"strings"
)

var emojis = []string{
	"😀", "😃", "😄", "😁", "😆",
	"😅", "😂", "🤣", "😊", "😇",
	"😍", "😎", "😜", "😝", "😋",
	"😚", "😘", "😗", "😙", "😏",
}

var emojiMap = map[string]string{
	":sweatsmile:": "😅",
	":smile:":      "😀",
	":grin:":       "😁",
	":kiss:":       "😘",
	":sunglasses:": "😎",
}

func emojiKeyboard(myApp fyne.App, msgEntry *widget.Entry) {
	myWindow := myApp.NewWindow("")
	myWindow.SetFixedSize(true)

	// Create a grid for the emojis
	emojiGrid := container.New(layout.NewGridLayoutWithColumns(5))

	for _, e := range emojis {
		emoji := e // capture range variable
		emojiButton := widget.NewButton(emoji, func() {
			msgEntry.Text = msgEntry.Text + emoji
			msgEntry.Refresh()
		})
		emojiGrid.Add(emojiButton)
	}

	myWindow.SetContent(emojiGrid)
	myWindow.Show()
}

func refreshEmojis(input string) string {
	if !strings.Contains(input, ":") {
		return input
	}

	output := input
	for k, v := range emojiMap {
		output = strings.Replace(output, k, v, -1)
	}
	return output
}
