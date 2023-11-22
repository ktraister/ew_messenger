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
	"github.com/sirupsen/logrus"
	"image/color"
	"net/http"
	"strings"
	"time"
)

//this is how you show dialog box
//dialog.ShowConfirm("foo", "foo", nil, myWindow)

//different layouts avail
//https://developer.fyne.io/explore/layouts.html#border

var users = []string{}
var targetUser = ""
var stashedMessages = []Post{}
var globalConfig Configurations
var proxyMsgChan = make(chan string)

//okay you can optimize it
func cnv(input float64) float64 {
    switch input {
    case 0:
	return 1.0
    case -1:
	return .9
    case -2:
	return .8
    case -3:
	return .7
    case -4:
	return .6
    case -5:
	return .5
    case -6:
	return .4
    case -7:
	return .3
    case -8:
	return .2
    case -9:
	return .1
    case -10:
	return 0.0
    }
    return 0.0
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

func postStashedMessages() {
	for index, message := range stashedMessages {
		if message.User == targetUser {
			incomingMsgChan <- message
			plusOne := index + 1
			//index error here b/c of slice length
			if len(stashedMessages) >= 2 {
				stashedMessages = append(stashedMessages[:index], stashedMessages[plusOne:]...)
			} else {
				stashedMessages = []Post{}
			}
		}
	}
}

func messageStashed(user string) bool {
	for _, message := range stashedMessages {
		if message.User == user {
			return true
		}
	}
	return false
}

// this thread manages proxy status and symbols
func proxyMgr(logger *logrus.Logger, pStatus *widget.Label) {
	for {
		//we don't care what it says, just that it was added
		_ = <-proxyMsgChan
		logger.Debug("received tap")
		prxy := "default"
		//manage the indicator text/color
		switch pStatus.Text {
		case "Proxy Off":
			pStatus.Text = "Starting Proxy..."
			pStatus.Importance = widget.MediumImportance
			prxy = "up"
		case "Proxy Up!":
			pStatus.Text = "Stopping Proxy..."
			pStatus.Importance = widget.MediumImportance
			prxy = "down"
		case "Starting Proxy...":
			pStatus.Text = "Stopping Proxy..."
			pStatus.Importance = widget.MediumImportance
			prxy = "down"
		default:
			pStatus.Text = "Proxy Off"
			pStatus.Importance = widget.LowImportance
		}
		pStatus.Refresh()

		if prxy == "up" {
			go proxy(globalConfig, logger, pStatus)
			//sleep to give the os a chance to assign us a port and listen
			time.Sleep(1 * time.Second)
			//when we want our threads to read in new config
			globalConfig.RandomURL = fmt.Sprintf("https://localhost:%d/api/otp", proxyPort)
			globalConfig.ExchangeURL = fmt.Sprintf("wss://localhost:%d/ws", proxyPort)
			logger.Debug(globalConfig.RandomURL)
			logger.Debug(globalConfig.ExchangeURL)
			logger.Debug("proxy Up")
		} else if prxy == "down" {
			quit <- true
			globalConfig.RandomURL = configuredRandomURL
			globalConfig.ExchangeURL = configuredExchangeURL
			logger.Debug("proxy Down")
		}
	}
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

// send needs to be a wrapper thread for go functions
func send(logger *logrus.Logger, sendButton *widget.Button, progressBar *widget.ProgressBarInfinite, textBox *widget.Entry) {
	for {
		message := <-outgoingMsgChan
		//set container to sending progressbar widget
		sendButton.Hide()
		textBox.Hide()
		progressBar.Show()
		//set container to sending progressbar widget

		//update user and send message to server root socket
		ok := ew_client(logger, globalConfig, message)
		//reset container to prior
		sendButton.Show()
		textBox.Show()
		progressBar.Hide()

		//post our sent message
		incomingMsgChan <- Post{Msg: message.Msg, User: globalConfig.User, ok: ok}
	}
}

func post(container *fyne.Container) {
	for {
		message := <-incomingMsgChan
		//this approach works
		if message.User == "Sending messages to" {
			messageLabel := widget.NewLabelWithStyle("Sending messages to: "+message.Msg, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			container.Add(messageLabel)
		} else if targetUser != message.User && globalConfig.User != message.User {
			//we're not focused on the user the message is from
			//stash the message for now
			stashedMessages = append(stashedMessages, message)
		} else if message.ok {
			messageLabel := widget.NewLabel(fmt.Sprintf("%s: %s", message.User, message.Msg))
			container.Add(messageLabel)
		} else {
			messageLabel := widget.NewLabel(fmt.Sprintf("ERROR SENDING MSG %s", message.Msg))
			messageLabel.Importance = widget.DangerImportance
			container.Add(messageLabel)
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

func configureGUI(myWindow fyne.Window, logger *logrus.Logger) {
	// Create a scrollable container for chat messages
	chatContainer := container.NewVBox()
	scrollContainer := container.NewVScroll(chatContainer)
	scrollContainer.Resize(fyne.NewSize(500, 0))

	//set greeting warning lable
	messageLabel := widget.NewLabelWithStyle("Select a user to send messages", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	messageLabel.Importance = widget.MediumImportance
	chatContainer.Add(messageLabel)

	// Create an entry field for typing messages
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Type your message...")
	//hiding the entry until a user is selected
	//come up with something cute to go here
	messageEntry.Hide()

	// add lines to use with onlinePanel
	text := widget.NewLabelWithStyle("    Online Users    ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	topLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	topLine2 := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	topLine.StrokeWidth = 5
	topLine2.StrokeWidth = 3
	bLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	bLine.StrokeWidth = 2
	sideLine := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	sideLine.StrokeWidth = 5
	sideLine2 := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	sideLine2.StrokeWidth = 5

	// add onlineUsers panel to show and select users
	onlineUsers := container.NewHBox(text)
	onlineUsers = container.NewBorder(nil, bLine, nil, sideLine2, onlineUsers)
	onlineUsers = container.NewBorder(onlineUsers, nil, nil, sideLine)

	//add a goroutine here to read ExchangeAPI for live users and populate with labels
	go refreshUsers(logger, onlineUsers)

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
		messageEntry.Show()
		targetUser = users[id]
		//clear the chat when switching users
		chatContainer.Objects = chatContainer.Objects[:0]
		messageLabel.Hide()
		chatContainer.Refresh()
		incomingMsgChan <- Post{Msg: users[id], User: "Sending messages to", ok: true}
		postStashedMessages()
		messageEntry.SetText("")
	}

	//actually add the users to the panel
	onlineUsers.Add(userList)
	//add container to hold the users list
	onlineContainer := container.New(layout.NewHBoxLayout(), onlineUsers)

	//define the sendbutton and OnClickFunc
	sendButton := widget.NewButton("Send", func() {
		// Get the message text from the entry field
		message := messageEntry.Text
		if message != "" {
			//check, spelled like it sounds
			if targetUser == globalConfig.User {
				incomingMsgChan <- Post{Msg: "Sending messages to yourself is not allowed", User: "SYSTEM", ok: false}
				return
			}

			//drop the messsage on the outgoing channel
			outgoingMsgChan <- Post{Msg: message, User: targetUser, ok: true}

			// Clear the message entry field after sending
			messageEntry.SetText("")
		}
	})
	//turn the send button blue
	sendButton.Importance = widget.HighImportance

	//define progress bar to use when sending a message
	infinite := widget.NewProgressBarInfinite()
	buttonContainer := container.New(layout.NewVBoxLayout(), infinite)
	buttonContainer.Add(sendButton)
	infinite.Hide()

	//define the chat clear button
	clearButton := widget.NewButton("Clear", func() {
		//clear chatContainer and messageEntry
		chatContainer.Objects = chatContainer.Objects[:0]
		chatContainer.Refresh()
		messageEntry.SetText("")
	})
	clearButton.Importance = widget.DangerImportance

	//create the widget to display current user
	myText := widget.NewLabelWithStyle("Logged in as: "+globalConfig.User, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	myText.Importance = widget.WarningImportance
	textContainer := container.New(layout.NewCenterLayout(), myText)

	//create proxy status widget
	pStatus := widget.NewLabelWithStyle("Proxy Off", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	pStatus.Importance = widget.LowImportance

	//create interactive proxy button
	proxyButton := widget.NewButton("Proxy", func() {
		logger.Debug("tapped")
		proxyMsgChan <- ""
	})

	//toolbar
	volp := widget.NewProgressBar()
	volp.SetValue(cnv(volume))
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.HelpIcon(), func() {
		        logger.Debug("help!")
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.VolumeUpIcon(), func() {
			if volume == 0 {
				return
			}
			volume += 1
			volp.SetValue(cnv(volume))
			logger.Debug(volume)
		}),
		widget.NewToolbarAction(theme.VolumeDownIcon(), func() {
			if volume == -10 {
				return
			}
			volume -= 1
			volp.SetValue(cnv(volume))
			logger.Debug(volume)
		}),
		widget.NewToolbarSpacer(),
	)
	//alert selection
	alerts := []string{"warning_beep", "navi_listen"}
	alertSelect := widget.NewSelect(alerts, func(input string){
		logger.Debug(input)
                selectedSound = input
	})
	toolBarContainer := container.NewBorder(nil, nil, nil, volp, toolbar)
	toolBarContainer = container.NewBorder(nil, nil, nil, alertSelect, toolBarContainer)


	//create container to hold current user/proxy button
	topContainer := container.NewHBox()
	sideLine3 := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	sideLine3.StrokeWidth = 5
	sideLine4 := canvas.NewLine(color.RGBA{0, 0, 0, 255})
	sideLine4.StrokeWidth = 2
	topContainer = container.NewBorder(nil, nil, nil, sideLine3, textContainer)
	topContainer = container.NewBorder(nil, nil, nil, pStatus, topContainer)
	topContainer = container.NewBorder(nil, nil, nil, sideLine4, topContainer)
	topContainer = container.NewBorder(nil, nil, nil, proxyButton, topContainer)

	// Create a container for the message entry container, clear button widget and send button container
	sendContainer := container.NewBorder(clearButton, buttonContainer, nil, nil, messageEntry)

	// Create a vertical split container for chat and input
	splitContainer := container.NewVSplit(scrollContainer, sendContainer)
	splitContainer.Offset = .7
	//Create borders for buttons
	finalContainer := container.NewBorder(topLine, nil, onlineContainer, nil, splitContainer)
	finalContainer = container.NewBorder(topContainer, nil, nil, nil, finalContainer)
	finalContainer = container.NewBorder(topLine2, nil, nil, nil, finalContainer)
	finalContainer = container.NewBorder(toolBarContainer, nil, nil, nil, finalContainer)

	//replace button in buttonContainer with progressBar when firing message
	//https://developer.fyne.io/widget/progressbar
	//listen for incoming messages here
	go listen(logger)
	go proxyMgr(logger, pStatus)
	go send(logger, sendButton, infinite, messageEntry)
	go post(chatContainer)

	myWindow.SetContent(finalContainer)
	myWindow.Resize(fyne.NewSize(600, 800))
	myWindow.Show()
}

func afterLogin(logger *logrus.Logger, myApp fyne.App, loginWindow fyne.Window) {
	//myApp.Preferences().SetString("AppTimeout", string(time.Minute))
	myWindow := myApp.NewWindow("EW Messenger")
	myWindow.SetMaster()
	configureGUI(myWindow, logger)
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
			afterLogin(logger, myApp, w)
			w.Close()
		}, w)
	}))
	w.RequestFocus()
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(450, 300))
	w.Show()
	myApp.Run()
}
