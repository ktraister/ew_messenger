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

type URI struct {
    val string
}

func (re URI) Authority() string {
    val := strings.Split(re.val, "//")[1]
    val = strings.Split(val, "/")[0]
    return val
}

func (re URI) Extension() string {
    return "Extension"
}

func (re URI) Fragment() string {
    if !strings.Contains(re.val, "#") {
	return ""
    }
    return strings.Split(re.val, "#")[1]

}

func (re URI) MimeType() string {
    return "MimeType"
}

func (re URI) Name() string {
    return "Name"
}

func (re URI) Path() string {
    val := strings.Split(re.val, "/")[3:]
    val2 := strings.Join(val, "")
    return "/" + val2
}

func (re URI) String() string {
    return "String"
}

func (re URI) Query() string {
    if !strings.Contains(re.val, "?") {
	return ""
    }
    return strings.Split(re.val, "?")[1]
}

func (re URI) Scheme() string {
    return strings.Split(re.val, ":")[0]
}

