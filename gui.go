package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
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
	"net/http"
	"net/url"
	"strings"
	"time"
)

//this is how you show dialog box
//dialog.ShowConfirm("foo", "foo", nil, myWindow)

//different layouts avail
//https://developer.fyne.io/explore/layouts.html#border

var users = []string{}
var targetUser = ""
var globalConfig Configurations
var stashedMessages = syncmap.Map{}
var chanMap = syncmap.Map{}

// okay you can optimize it
func cnv(input float64) float64 {
	factor := -(input)
	return 1.0 - 0.1*factor
}

func checkCreds() (bool, string) {
	//setup tls
	ts := tlsClient(globalConfig.RandomURL)
	//check and make sure inserted creds
	//Random and Exchange will use same mongo, so the creds will be valid for both
	health_url := fmt.Sprintf("%s%s", globalConfig.RandomURL, "/healthcheck")
	req, err := http.NewRequest("GET", health_url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", globalConfig.User)
	req.Header.Set("Passwd", globalConfig.Passwd)
	client := http.Client{Timeout: 3 * time.Second, Transport: ts}
	resp, err := client.Do(req)
	errorText := ""
	if err != nil {
		errorText = "Couldn't Connect to RandomAPI"
		fmt.Println(errorText + " " + globalConfig.RandomURL)
		fmt.Println("Quietly exiting now. Please reconfigure.")
		return false, errorText
	}
	if resp == nil {
		errorText = "No Response From RandomAPI"
		fmt.Println(errorText + " " + globalConfig.RandomURL)
		fmt.Println("Quietly exiting now. Please reconfigure.")
		return false, errorText
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errorText = "Invalid Username/Password Combination"
		fmt.Println(errorText)
		fmt.Printf("Request failed with status: %s\n", resp.Status)
		return false, errorText
	}
	return true, ""
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
		_, incoming, err := cm.Read()
		if err != nil {
			logger.Error("Error reading message:", err)
			continue
		}

		err = json.Unmarshal([]byte(incoming), &dat)
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
			incomingMsgChan <- Post{Msg: "Sending messages to yourself is not allowed", From: "SYSTEM", To: "SYSTEM", ok: false}
			return
		}

		//drop the messsage on the outgoing channel
		outgoingMsgChan <- Post{Msg: message, To: targetUser, From: globalConfig.User, ok: true}

		// Clear the message entry field after sending
		messageEntry.SetText("")
	}

}

func send(logger *logrus.Logger, textBox *widget.Entry) {
	for {
		message := <-outgoingMsgChan

		//update user and send message
		ok := ew_client(logger, globalConfig, message)

		//post our sent message
		incomingMsgChan <- Post{Msg: message.Msg, To: message.To, From: globalConfig.User, ok: ok}
	}
}

// okay fuck it we're calling the text boxes good for now
func post(cont *fyne.Container, userChan chan Post) {
	for {
		line := canvas.NewLine(color.RGBA{255, 255, 255, 20})
		line.StrokeWidth = 0.2
		message := <-userChan
		if message.ok {
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
			messageLabel := widget.NewLabel(fmt.Sprintf("ERROR SENDING MSG %s", message.Msg))
			messageLabel.Importance = widget.DangerImportance
			cont.Add(messageLabel)
			cont.Add(line)
		}
	}
}

func refreshUsers(logger *logrus.Logger, container *fyne.Container) {
	for {
		users = []string{}
		users, _ = getExUsers(logger, globalConfig)
		//logger.Debug("refreshUsers --> ", users)
		container.Refresh()
		//refresh rate
		time.Sleep(1 * time.Second)
	}
}

func afterLogin(logger *logrus.Logger, myApp fyne.App) {
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
	onlineUsers := container.NewMax()

	//add a goroutine here to read ExchangeAPI for live users and populate with labels
	go refreshUsers(logger, onlineUsers)

	//goroutine to route messages
	go msgRouter(logger)

	//build our user list
	userList := widget.NewList(
		//length
		func() int {
			return len(users)
		},
		//create Item
		func() fyne.CanvasObject {
			label := widget.NewLabel("Text")
			return container.NewBorder(nil, nil, nil, nil, label)
		},
		//updateItem
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			text := obj.(*fyne.Container).Objects[0].(*widget.Label)
			text.SetText(users[id])
			if messageStashed(users[id]) {
				//turn the user blue if we have messages from them
				text.Importance = widget.HighImportance
			} else {
				//reset user text
				text.Importance = widget.MediumImportance
			}
		})
	userList.OnSelected = func(id widget.ListItemID) {
		//setting global scoped var
		targetUser = users[id]
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

	onlineUsers.Add(userList)

	//add container to hold the users list
	bLine2 := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	bLine2.StrokeWidth = 2

	//create the widget to display current user
	userText := widget.NewLabelWithStyle("Online Users", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	myText := widget.NewLabelWithStyle(globalConfig.User, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	myText.Importance = widget.WarningImportance
	textContainer := container.New(layout.NewCenterLayout(), myText)
	uTextContainer := container.New(layout.NewCenterLayout(), userText)

	//create proxy status widget
	pStatus := widget.NewLabelWithStyle("Starting Proxy...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	pStatus.Importance = widget.LowImportance
	go proxy(globalConfig, logger, pStatus)

	//toolbar
	volp := widget.NewProgressBar()
	volp.SetValue(cnv(volume))
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			logger.Debug("help!")
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
	topContainer = container.NewBorder(nil, nil, nil, pStatus, topContainer)
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

func main() {
	//globalConfig stuff
	globalConfig = fetchConfig()
	logger := createLogger(globalConfig.LogLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("randomURL is\t\t", globalConfig.RandomURL)
	logger.Debug("exchangeURL is\t", globalConfig.ExchangeURL)
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

			//create our hasher to hash our pass
			hash := sha512.New()
			hash.Write([]byte(password.Text))
			hashSum := hash.Sum(nil)
			hashString := hex.EncodeToString(hashSum)

			//set values we just took in with login widget
			globalConfig.User = strings.ToLower(username.Text)
			globalConfig.Passwd = hashString
			logger.Debug(hashString)

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
