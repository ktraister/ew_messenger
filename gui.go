package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/syncmap"
	"image/color"
	"net/url"
	"strings"
	"time"
)

var activeUsers = []string{}
var friendUsers = []string{}
var nonFriendUsers = []string{}
var tmpFriendUsers = []string{}
var targetUser = ""
var globalConfig Configurations
var stashedMessages = syncmap.Map{}
var chanMap = syncmap.Map{}

var helpLock = false
var statusLock = false
var friendsLock = false

// values used to display system status
var warningCont = container.NewVBox()
var statusButton = widget.NewButton("Status", func() {})
var aStatus = widget.NewLabelWithStyle("GO", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
var eStatus = widget.NewLabelWithStyle("GO", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
var mStatus = widget.NewLabelWithStyle("GO", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
var pStatus = widget.NewLabelWithStyle("GO", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

type statusMsg struct {
	Target string
	Text   string
	Import widget.Importance
	Warn   string
}

var statusMsgChan = make(chan statusMsg)

// okay you can optimize it
func cnv(input float64) float64 {
	factor := -(input)
	return 1.0 - 0.1*factor
}

func removeFriends() {
	for _, user := range activeUsers {
		if !isFriend(user) && !isNonFriend(user) {
			nonFriendUsers = append(nonFriendUsers, user)
		}
	}
}

func isActive(user string) bool {
	for _, element := range activeUsers {
		if element == user {
			return true
		}
	}
	return false
}

func isFriend(user string) bool {
	for _, element := range friendUsers {
		if element == user {
			return true
		}
	}
	return false
}

func isTmpFriend(user string) bool {
	for _, element := range tmpFriendUsers {
		if element == user {
			return true
		}
	}
	return false
}

func isNonFriend(user string) bool {
	for _, element := range nonFriendUsers {
		if element == user {
			return true
		}
	}
	return false
}

// this function will route messages from incoming to correct chan
func msgRouter(logger *logrus.Logger) {
	for {
		message := <-incomingMsgChan
		//route the message
		toFlag := false
		fromFlag := false
		var v interface{}
		v, ok := chanMap.Load(message.To)
		if !ok {
			toFlag = true
			v, ok = chanMap.Load(message.From)
			if !ok {
				fromFlag = true
			}
		}
		if toFlag && fromFlag {
			//stash the message
			logger.Debug("Stashing msg ", message)
			stashedMessages.Store(uuid.New().String(), message)
			continue
		}
		logger.Debug("Posting message ", message)
		ch := v.(chan Post)
		ch <- message
	}
}

// function to control status widgets
func statusMgr(logger *logrus.Logger) {
	allowStatusUpdate := true
	for {
		message := <-statusMsgChan

		if allowStatusUpdate {
			switch message.Import {
			case widget.WarningImportance:
				statusButton.Importance = widget.WarningImportance
			case widget.DangerImportance:
				statusButton.Importance = widget.DangerImportance
			default:
				statusButton.Importance = widget.SuccessImportance
			}
			statusButton.Refresh()
		}

		myT := message.Target
		switch myT {
		case "API":
			aStatus.Importance = message.Import
			aStatus.Text = message.Text
			aStatus.Refresh()
		case "EX":
			eStatus.Importance = message.Import
			eStatus.Text = message.Text
			eStatus.Refresh()
		case "MITM":
			mStatus.Importance = message.Import
			mStatus.Text = message.Text
			mStatus.Refresh()
		case "PROXY":
			pStatus.Importance = message.Import
			pStatus.Text = message.Text
			pStatus.Refresh()
			allowStatusUpdate = false
		}

		if message.Warn != "" {
			if message.Import == widget.LowImportance {
				message.Import = widget.MediumImportance
			}
			newWidget := widget.NewLabel(message.Warn)
			newWidget.Importance = message.Import
			warningCont.Add(newWidget)

		}
	}
}

func postStashedMessages(targetUser string) {
	stashedMessages.Range(func(key, value interface{}) bool {
		// cast value to correct format
		msg, ok := value.(Post)
		if !ok {
			// this will break iteration
			return false
		}

		if msg.From == targetUser {
			ch, ok := chanMap.Load(msg.From)
			if !ok {
				return false
			}
			channel := ch.(chan Post)
			//send and remove
			channel <- msg
			stashedMessages.Delete(key)
		}

		// this will continue iterating
		return true
	})
}

// if we don't have a chan, return false
func messageStashed(user string) bool {
	_, ok := chanMap.Load(user)
	if !ok {
		flag := false
		//we dont have a chan, therefor we need to check
		stashedMessages.Range(func(key, value interface{}) bool {
			msg, _ := value.(Post)
			if msg.From == user {
				flag = true
				return false
			}
			// this will continue iterating
			return true
		})
		if flag {
			return true
		} else {
			return false
		}
	}
	return false
}

// this thread should just read HELO and pass off to another thread
func listen(logger *logrus.Logger) {
	localUser := fmt.Sprintf("%s_%s", globalConfig.User, "server")
	cm, err := exConnect(logger, globalConfig, localUser)
	if err != nil {
		return
	}
	defer cm.Close()

	//listen for incoming connections
	for {
		incoming, err := cm.Read()
		if err != nil {
			logger.Error("Error reading message:", err)
			continue
		}

		err = json.Unmarshal(incoming, &dat)
		if err != nil {
			logger.Error("Error unmarshalling json:", err)
			continue
		}

		//new connections should always ask
		if dat["msg"] == "HELO" {
			logger.Debug("Received HELO from ", dat["from"])
		} else {
			logger.Warn("New connection didn't HELO, bouncing")
			continue
		}
		//handle connection creates new socket inside goRoutine
		go handleConnection(dat, logger, globalConfig)
	}
}

// function to send message from GUI
func sendMsg(messageEntry *widget.Entry) {
	// Get the message text from the entry field
	message := messageEntry.Text
	if message != "" {
		//check, spelled like it sounds
		if targetUser == globalConfig.User {
			incomingMsgChan <- Post{Msg: "Sending messages to yourself is not allowed", From: "SYSTEM", To: "SYSTEM", Err: errors.New("SYSTEM: Self-sending is not allowed")}
			return
		}

		//drop the messsage on the outgoing channel
		outgoingMsgChan <- Post{Msg: message, To: targetUser, From: globalConfig.User}

		// Clear the message entry field after sending
		messageEntry.SetText("")
	}

}

func send(logger *logrus.Logger, textBox *widget.Entry) {
	for {
		message := <-outgoingMsgChan

		//update user and send message
		err := ew_client(logger, globalConfig, message)

		//post our sent message
		if err == nil {
			incomingMsgChan <- Post{Msg: message.Msg, To: message.To, From: globalConfig.User}
		} else {
			incomingMsgChan <- Post{Msg: message.Msg, To: message.To, From: globalConfig.User, Err: err}
		}
	}
}

// okay fuck it we're calling the text boxes good for now
func post(cont *fyne.Container, userChan chan Post) {
	for {
		line := canvas.NewLine(color.RGBA{255, 255, 255, 20})
		line.StrokeWidth = 0.2
		message := <-userChan
		if message.Err == nil {
			//regex is misbehaving rn
			u, err := url.Parse(message.Msg)
			if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
				linkLabel := widget.NewHyperlinkWithStyle(message.Msg, u, fyne.TextAlignTrailing, fyne.TextStyle{})
				if message.From == globalConfig.User {
					linkLabel.Alignment = fyne.TextAlignLeading
				}
				cont.Add(linkLabel)
				cont.Add(line)
			} else {
				//plaintext
				messageLabel := widget.NewLabelWithStyle(fmt.Sprintf("%s", message.Msg), fyne.TextAlignTrailing, fyne.TextStyle{})
				if message.From == globalConfig.User {
					messageLabel.Alignment = fyne.TextAlignLeading
				}
				messageLabel.Wrapping = fyne.TextWrapWord
				cont.Add(messageLabel)
				cont.Add(line)
			}
		} else {
			messageLabel := widget.NewLabel(fmt.Sprintf("%s", message.Err))
			messageLabel.Importance = widget.DangerImportance
			messageLabel.Wrapping = fyne.TextWrapWord
			cont.Add(messageLabel)
			cont.Add(line)
		}
	}
}

func refreshUsers(logger *logrus.Logger, userContainer *fyne.Container, friendContainer *fyne.Container) {
	for {
		friendUsers, _ = getFriends(logger)
		friendContainer.Refresh()
		activeUsers, _ = getExUsers(logger)
		removeFriends()
		userContainer.Refresh()
		//refresh rate
		time.Sleep(1 * time.Second)
	}
}

func afterLogin(logger *logrus.Logger, myApp fyne.App) {
	//goroutine to route messages
	go msgRouter(logger)

	//goroutines to check for api and exchange status
	go apiStatusCheck(logger)
	go exStatusCheck(logger)
	go mitmStatusCheck(logger)

	//statusManager goroutine
	go statusMgr(logger)

	//setup New window
	myWindow := myApp.NewWindow("EW Messenger")
	myWindow.SetMaster()

	// add lines to use with onlinePanel
	topLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	topLine.StrokeWidth = 1
	bLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	bLine.StrokeWidth = 5
	sideLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	sideLine.StrokeWidth = 3

	// add onlineUsers panel to show and select users
	//build our user list
	userList := widget.NewList(
		//length
		func() int {
			return len(nonFriendUsers)
		},
		//create Item
		func() fyne.CanvasObject {
			label := widget.NewLabel("Text")
			return container.NewBorder(nil, nil, nil, nil, label)
		},
		//updateItem
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			text := obj.(*fyne.Container).Objects[0].(*widget.Label)
			text.SetText(nonFriendUsers[id])
			if messageStashed(nonFriendUsers[id]) {
				//turn the user blue if we have messages from them
				text.Importance = widget.HighImportance
			} else {
				//reset user text
				text.Importance = widget.MediumImportance
			}
		})
	userList.OnSelected = func(id widget.ListItemID) {
		//setting global scoped var
		targetUser = nonFriendUsers[id]
		//dont show as selected
		userList.UnselectAll()

		//create the new chan for the user here if not exists
		_, ok := chanMap.Load(targetUser)
		if ok {
			return
		}
		ch := make(chan Post)
		chanMap.Store(targetUser, ch)
		newConvoWin(logger, myApp, targetUser, ch)
		postStashedMessages(targetUser)
	}

	// add friendList panel to show and select users
	//build our friends list
	friendList := widget.NewList(
		//length
		func() int {
			return len(friendUsers)
		},
		//create Item
		func() fyne.CanvasObject {
			label := widget.NewLabel("Text")
			return container.NewBorder(nil, nil, nil, nil, label)
		},
		//updateItem
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			text := obj.(*fyne.Container).Objects[0].(*widget.Label)
			text.SetText(friendUsers[id])
			if messageStashed(friendUsers[id]) {
				//turn the user blue if we have messages from them
				text.Importance = widget.HighImportance
			} else {
				//reset user text
				text.Importance = widget.MediumImportance
				if !isActive(friendUsers[id]) {
					text.Importance = widget.LowImportance
				}
			}
		})
	friendList.OnSelected = func(id widget.ListItemID) {
		//setting global scoped var
		targetUser = friendUsers[id]
		//dont show as selected
		friendList.UnselectAll()

		//create the new chan for the user here if not exists
		_, ok := chanMap.Load(targetUser)
		if ok {
			return
		}
		ch := make(chan Post)
		chanMap.Store(targetUser, ch)
		newConvoWin(logger, myApp, targetUser, ch)
		postStashedMessages(targetUser)
	}

	friendButton := widget.NewButton("Manage Friends", func() {
		manageFriendsWin(logger, myApp)
	})
	friendButton.Importance = widget.LowImportance
	friendText := widget.NewLabelWithStyle("Friends", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	userContainer := container.NewMax(userList)
	friendContainer := container.NewMax(friendList)
	friendContainer = container.NewBorder(friendText, friendButton, nil, nil, friendContainer)
	onlineUsers := container.NewVSplit(userContainer, friendContainer)
	onlineUsers.SetOffset(.6)

	//setUp friendUsers slices here
	friendUsers, _ = getFriends(logger)
	tmpFriendUsers = friendUsers
	//add a goroutine here to read ExchangeAPI for live users and populate with labels
	go refreshUsers(logger, userContainer, friendContainer)

	//add container to hold the users list
	bLine2 := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	bLine2.StrokeWidth = 2

	//create the widget to display current user
	userText := widget.NewLabelWithStyle("Online Users", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	myText := widget.NewLabelWithStyle(globalConfig.User, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	myText.Importance = widget.WarningImportance
	textContainer := container.New(layout.NewCenterLayout(), myText)
	uTextContainer := container.New(layout.NewCenterLayout(), userText)

	//create status button -- defined line 37
	statusButton = widget.NewButton("Status", func() { systemStatus(myApp) })
	statusButton.Importance = widget.SuccessImportance
	go proxy(logger)

	//toolbar
	volp := widget.NewProgressBar()
	volp.SetValue(cnv(volume))
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			help(myApp)
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.VolumeDownIcon(), func() {
			if volume == -10 {
				return
			}
			volume -= 1
			volp.SetValue(cnv(volume))
			logger.Debug(volume)
		}),
		widget.NewToolbarAction(theme.VolumeUpIcon(), func() {
			if volume == 0 {
				return
			}
			volume += 1
			volp.SetValue(cnv(volume))
			logger.Debug(volume)
		}),
		widget.NewToolbarSpacer(),
	)
	//alert selection
	alerts := []string{"warning_beep", "navi_listen", "off"}
	alertSelect := widget.NewSelect(alerts, func(input string) {
		logger.Debug(input)
		selectedSound = input
	})
	alertSelect.SetSelected("navi_listen")
	toolBarContainer := container.NewBorder(nil, nil, nil, volp, toolbar)
	toolBarContainer = container.NewBorder(nil, nil, nil, alertSelect, toolBarContainer)

	//create container to hold current user/proxy
	topContainer := container.NewHBox()
	topContainer = container.NewBorder(nil, nil, nil, sideLine, textContainer)
	topContainer = container.NewBorder(nil, nil, nil, statusButton, topContainer)
	topContainer = container.NewBorder(nil, bLine2, nil, nil, topContainer)
	topContainer = container.NewBorder(nil, uTextContainer, nil, nil, topContainer)

	//Create borders for buttons
	finalContainer := container.NewBorder(bLine, nil, nil, nil, nil)
	finalContainer.Add(onlineUsers)
	finalContainer = container.NewBorder(topContainer, nil, nil, nil, finalContainer)
	finalContainer = container.NewBorder(topLine, nil, nil, nil, finalContainer)
	finalContainer = container.NewBorder(toolBarContainer, nil, nil, nil, finalContainer)

	//https://developer.fyne.io/widget/progressbar
	//listen for incoming messages here
	go listen(logger)

	myWindow.SetContent(finalContainer)
	myWindow.Resize(fyne.NewSize(200, 600))
	myWindow.Show()

	if !binIsCurrent(logger) {
		logger.Debug("Bin is not current!")
		verWindow := myApp.NewWindow("Important Information")
		verWindow.CenterOnScreen()
		cont := container.NewVBox()
		cont.Add(widget.NewLabel("A new version of the EW_Messenger is available."))
		u, _ := url.Parse("https://endlesswaltz.xyz/downloads")
		linkLabel := widget.NewHyperlinkWithStyle("Download the new version here.", u, fyne.TextAlignCenter, fyne.TextStyle{})
		cont.Add(linkLabel)
		verWindow.SetContent(cont)
		verWindow.Show()
		verWindow.RequestFocus()

	} else {
		logger.Debug("Bin IS current!")
	}
}

func help(myApp fyne.App) {
	if helpLock {
		return
	}

	helpLock = true

	myWindow := myApp.NewWindow("Help")

	helpCont := container.NewVBox()
	helpDisplayCont := container.NewVScroll(helpCont)
	helpText := widget.NewLabelWithStyle("EW Messenger Help", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	helpCont.Add(helpText)
	mainWindow := widget.NewLabelWithStyle("The Main Window", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	helpCont.Add(mainWindow)
	topLine := widget.NewLabelWithStyle(`The top line of the main messenger window allows for sound configuration. From left to right, the buttons allow you to lower and raise volume; volume display, and alert sound selection. You can also choose to mute the application in the sound dropdown`, fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	topLine.Wrapping = fyne.TextWrapWord
	helpCont.Add(topLine)

	line2 := widget.NewLabelWithStyle(`The second line displays the logged in user's username, next to a status button. This button will change colors depending on the status of the messenger and messenger infrastructure. Clicking the button will show a status window, displaying granular system status and any error messages.`, fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	line2.Wrapping = fyne.TextWrapWord
	helpCont.Add(line2)

	line3 := widget.NewLabelWithStyle(`The "Online Users" portion of the panel shows all users currently logged on to the exchange. A user who has sent you a message will appear blue, while any other user will appear white.`, fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	line3.Wrapping = fyne.TextWrapWord
	helpCont.Add(line3)

	line4 := widget.NewLabelWithStyle(`The "Friends" portion of the panel shows all users you have added to your friends list. To manage your friends list, click the "Manage Friends" button at the bottom of the main panel. Another window will open, allowing you to select or deselect any existing user to be part of your friends list. Any of your friends who has sent you a message will appear to be blue.`, fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	line4.Wrapping = fyne.TextWrapWord
	helpCont.Add(line4)

	messengerWindow := widget.NewLabelWithStyle("Sending Messages", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	helpCont.Add(messengerWindow)
	line5 := widget.NewLabelWithStyle(`To send a message, click on your target username. A new window will open, allowing you to send message to and receive messages from the target user.`, fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	line5.Wrapping = fyne.TextWrapWord
	helpCont.Add(line5)

	line6 := widget.NewLabelWithStyle(`When you close the message window, you will lose your current chat history.`, fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	line6.Wrapping = fyne.TextWrapWord
	helpCont.Add(line6)

	line7 := widget.NewLabelWithStyle(`To send messages without hitting "Send", use the combination "Shift+Enter". Emojis can be entered with the emoji keyboard button, or by typing the emoji name - EX:":grin:"`, fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	line7.Wrapping = fyne.TextWrapWord
	helpCont.Add(line7)

	myWindow.SetContent(helpDisplayCont)
	myWindow.SetFixedSize(true)
	myWindow.Resize(fyne.NewSize(300, 300))
	myWindow.SetOnClosed(func() {
		helpLock = false
	})
	myWindow.Show()
}

func systemStatus(myApp fyne.App) {
	if statusLock {
		return
	}

	statusLock = true

	myWindow := myApp.NewWindow("System Status")

	line0 := canvas.NewLine(color.RGBA{255, 255, 255, 20})
	line0.StrokeWidth = 0.2
	line1 := canvas.NewLine(color.RGBA{255, 255, 255, 20})
	line1.StrokeWidth = 0.2
	line2 := canvas.NewLine(color.RGBA{255, 255, 255, 20})
	line2.StrokeWidth = 0.2

	systemCont := container.NewHBox()
	sysGrid := container.New(layout.NewGridLayoutWithColumns(2))

	header := widget.NewLabelWithStyle("Status", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	systemCont.Add(header)

	aStatus.Importance = widget.SuccessImportance
	sysGrid.Add(aStatus)
	aText := widget.NewLabelWithStyle("API", fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	sysGrid.Add(aText)

	eStatus.Importance = widget.SuccessImportance
	sysGrid.Add(eStatus)
	eText := widget.NewLabelWithStyle("Exchange", fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	sysGrid.Add(eText)

	sysGrid.Add(mStatus)
	mText := widget.NewLabelWithStyle("MITM", fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	sysGrid.Add(mText)

	sysGrid.Add(pStatus)
	pText := widget.NewLabelWithStyle("Proxy", fyne.TextAlignLeading, fyne.TextStyle{Bold: false})
	sysGrid.Add(pText)

	sysCont := container.NewHBox()
	sysCont.Add(sysGrid)

	warnCont := container.NewHBox()
	warnDisplayCont := container.NewVScroll(warningCont)
	warnCont.Add(warnDisplayCont)
	warnText := widget.NewLabelWithStyle("Warnings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	warnCont = container.NewBorder(line2, nil, nil, nil, warnCont)
	warnCont = container.NewBorder(warnText, nil, nil, nil, warnCont)

	finalCont := container.NewBorder(nil, line0, nil, nil, systemCont)
	finalCont = container.NewBorder(nil, sysCont, nil, nil, finalCont)
	finalCont = container.NewBorder(nil, nil, nil, line1, finalCont)
	finalCont = container.NewBorder(nil, nil, nil, warnCont, finalCont)
	myWindow.SetContent(finalCont)
	myWindow.SetFixedSize(true)
	myWindow.Resize(fyne.NewSize(400, 100))
	myWindow.SetOnClosed(func() {
		statusLock = false
	})
	myWindow.Show()
}

func newConvoWin(logger *logrus.Logger, myApp fyne.App, user string, userChan chan Post) {
	myWindow := myApp.NewWindow(user)

	// Create a scrollable container for chat messages
	chatContainer := container.NewVBox()
	scrollContainer := container.NewVScroll(chatContainer)

	// Create an entry field for typing messages
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Type your message -- Shift + Enter to send")
	messageEntry.Wrapping = fyne.TextWrapWord

	//replace input with emojis
	messageEntry.OnChanged = func(input string) {
		messageEntry.Text = refreshEmojis(input)
		messageEntry.Refresh()
	}

	//WONT TRAP ENTER to send the message if rest of the gui in focus
	//https://github.com/fyne-io/fyne/issues/1683#issuecomment-755390386
	//this sends a message if shift+enter is pressed in focus
	//apparently people like this behaviour /shrug
	messageEntry.OnSubmitted = func(input string) {
		sendMsg(messageEntry)
	}

	//define the sendbutton and OnClickFunc
	sendButton := widget.NewButton("Send", func() { sendMsg(messageEntry) })
	//turn the send button blue
	sendButton.Importance = widget.HighImportance

	//emoji button
	emojiButton := widget.NewButton("Emojis", func() { emojiKeyboard(myApp, messageEntry) })

	// Create a container for the message entry container, clear button widget and send button container
	tmpContainer := container.NewBorder(nil, nil, nil, emojiButton, sendButton)
	msgContainer := container.NewVSplit(scrollContainer, messageEntry)
	msgContainer.Offset = .75
	sendContainer := container.NewBorder(nil, tmpContainer, nil, nil, msgContainer)

	go send(logger, messageEntry)
	go post(chatContainer, userChan)

	//close the channel when the window is closed
	myWindow.SetOnClosed(func() {
		close(userChan)
		chanMap.Delete(user)
	})

	myWindow.SetContent(sendContainer)
	myWindow.Resize(fyne.NewSize(350, 450))
	myWindow.Show()
}

func manageFriendsWin(logger *logrus.Logger, myApp fyne.App) {
	if friendsLock {
		return
	}

	friendsLock = true

	myWindow := myApp.NewWindow("Manage Friends")

	//allUsers []string{}
	allUsers, _ := getAllUsers(logger)
	//var cvObj *fyne.Container

	userList := widget.NewList(
		//length
		func() int {
			return len(allUsers)
		},
		//create Item
		func() fyne.CanvasObject {
			label := widget.NewLabel("Text")
			return container.NewBorder(nil, nil, nil, nil, label)
		},
		//updateItem
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			text := obj.(*fyne.Container).Objects[0].(*widget.Label)
			text.SetText(allUsers[id])
			if isTmpFriend(allUsers[id]) {
				text.Importance = widget.HighImportance
			} else {
				text.Importance = widget.MediumImportance
			}
		})
	userList.OnSelected = func(id widget.ListItemID) {
		if isTmpFriend(allUsers[id]) {
			var tmp []string
			for _, user := range tmpFriendUsers {
				if user != allUsers[id] {
					tmp = append(tmp, user)
				}
			}
			tmpFriendUsers = tmp

		} else {
			tmpFriendUsers = append(tmpFriendUsers, allUsers[id])
		}
		fmt.Println(friendUsers)
		userList.UnselectAll()
		userList.Refresh()
	}

	submitButton := widget.NewButton("Submit", func() {
		//POST new user list and close the window
		putFriends(logger)
		myWindow.Close()
	})
	activeText := widget.NewLabelWithStyle("Available Users", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	allContainer := container.NewMax(userList)
	allContainer = container.NewBorder(activeText, submitButton, nil, nil, allContainer)
	myWindow.SetContent(allContainer)
	myWindow.Resize(fyne.NewSize(250, 450))
	myWindow.SetOnClosed(func() {
		friendsLock = false
	})
	myWindow.Show()
}

func main() {
	//globalConfig stuff
	globalConfig = fetchConfig()
	logger := createLogger(globalConfig.LogLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("PrimaryURL is\t\t", globalConfig.PrimaryURL)
	logger.Debug("sshHost is\t\t", globalConfig.SSHHost)

	//add "starting up" message while loading
	myApp := app.NewWithID("EW Messenger")
	w := myApp.NewWindow("EW Messenger Login")
	username := widget.NewEntry()
	password := widget.NewPasswordEntry()
	content := widget.NewForm(widget.NewFormItem("Username", username),
		widget.NewFormItem("Password", password))
	w.SetContent(widget.NewButton("Login to the EW Circut", func() {
		dialog.ShowCustomConfirm("", "Log In", "Cancel", content, func(b bool) {
			logger.Debug("Checking creds...")

			//set values we just took in with login widget
			globalConfig.User = strings.ToLower(username.Text)
			globalConfig.Passwd = password.Text

			//pass the hash lol
			ok, err := checkCreds()

			if !ok || !b {
				//this is caused by an error in the checkCreds routine
				errorWidget := widget.NewLabel(err)
				errorWidget.Importance = widget.DangerImportance
				content = widget.NewForm(widget.NewFormItem("Login Error", errorWidget),
					widget.NewFormItem("Username", username),
					widget.NewFormItem("Password", password))
				return
			}
			logger.Debug("creds passed check!")

			//run the next window
			afterLogin(logger, myApp)
			w.Close()
		}, w)
	}))
	w.RequestFocus()
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(450, 300))
	w.Show()
	myApp.Run()
}
