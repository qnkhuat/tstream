/* Chat client on terminal */
package streamer

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/rivo/tview"
	"log"
	"math"
	"net/url"
	"strings"
	"time"
)

type Chat struct {
	username         string
	sessionId        string
	serverAddr       string
	color            string
	conn             *websocket.Conn
	app              *tview.Application
	startedTime      time.Time
	chatTextView     *tview.TextView
	nviewersTextView *tview.TextView
	uptimeTextView   *tview.TextView
	titleTextView    *tview.TextView
}

func NewChat(sessionId, serverAddr, username string) *Chat {
	return &Chat{
		username:   username,
		sessionId:  sessionId,
		serverAddr: serverAddr,
		color:      "red",
		app:        tview.NewApplication(),
	}
}

func (c *Chat) Start() error {
	c.initUI()

	err := c.connectWS()
	if err != nil {
		log.Printf("Error: %s", err)
	}

	// Receive
	go func() {
		for {
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("Failed to read message: %s", err)
				c.Stop()
				return
			}
			msgObj, err := message.Unwrap(msg)
			switch msgObj.Type {
			case message.TChat:
				var chatList []message.Chat
				err = json.Unmarshal(msgObj.Data, &chatList)
				if err != nil {
					log.Printf("Failed to decode chat message: %s", err)
					continue
				}
				c.addChatMsgs(chatList)
			case message.TRoomInfo:
				var roomInfo message.RoomInfo
				err = json.Unmarshal(msgObj.Data, &roomInfo)

				c.startedTime = roomInfo.StartedTime
				c.nviewersTextView.SetText(fmt.Sprintf("%d 👤", roomInfo.NViewers))
				c.titleTextView.SetText(fmt.Sprintf("%s", roomInfo.Title))

			default:
				log.Printf("Not implemented to handle message type: %s", msgObj.Type)

			}
		}
	}()

	c.requestServer(message.TRequestRoomInfo)
	c.requestServer(message.TRequestCacheChat)

	go func() {
		tick := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-tick.C:
				c.requestServer(message.TRequestRoomInfo)
			}
		}
	}()

	go func() {
		tick := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-tick.C:
				upTime := time.Since(c.startedTime)
				hours := int(math.Floor(upTime.Hours()))
				upTime = upTime - time.Duration(hours)*time.Hour

				minutes := int(math.Floor(upTime.Minutes()))
				upTime = upTime - time.Duration(minutes)*time.Minute

				seconds := int(math.Floor(upTime.Seconds()))

				upTimeStr := fmt.Sprintf("[red]%02d:%02d:%02d[white]", hours, minutes, seconds)
				c.app.QueueUpdateDraw(func() {
					c.uptimeTextView.SetText(upTimeStr)
				})

			}
		}
	}()

	if err := c.app.EnableMouse(true).Run(); err != nil {
		panic(err)
	}

	return nil
}

func (c *Chat) requestServer(msgType message.MType) error {
	reqMsg := message.Wrapper{
		Type: msgType,
		Data: []byte{},
	}
	payload, _ := json.Marshal(reqMsg)
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

func (c *Chat) initUI() error {
	layout := tview.NewGrid().
		SetRows(4, 0, 1).
		SetColumns(0).
		SetBorders(true)

	tstreamText := tview.NewTextView().
		SetText("TStream").
		SetTextAlign(tview.AlignCenter)

	usernameText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("@%s", c.username))

	c.titleTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetText("Title")

	c.nviewersTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		SetText("👤 10")

	c.uptimeTextView = tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetDynamicColors(true)

	header := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0, 0, 0).
		AddItem(tstreamText, 0, 0, 1, 3, 0, 0, false).
		AddItem(usernameText, 2, 0, 1, 2, 0, 0, false).
		AddItem(c.titleTextView, 1, 0, 1, 2, 0, 0, false).
		AddItem(c.nviewersTextView, 1, 2, 1, 1, 0, 0, false).
		AddItem(c.uptimeTextView, 2, 2, 1, 1, 0, 0, false)

	c.chatTextView = tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true).
		SetWordWrap(true).SetText("").
		ScrollToEnd()

	messageInput := tview.NewInputField()
	messageInput.SetLabel("[red]>[red] ").
		SetDoneFunc(func(key tcell.Key) {
			text := messageInput.GetText()
			if text == "/exit" {
				c.Stop()
				return
			}

			chat := message.Chat{
				Name:    c.username,
				Color:   c.color,
				Content: text,
				Time:    time.Now().String(),
				Role:    "Streamer",
			}

			chatList := []message.Chat{chat}
			msg, err := message.Wrap(message.TChat, chatList)

			if err == nil {
				payload, _ := json.Marshal(msg)
				c.conn.WriteMessage(websocket.TextMessage, payload)
			} else {
				log.Printf("Failed to wrap message")
			}
			c.addChatMsgs(chatList)
			messageInput.SetText("")
		})

	layout.AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(c.chatTextView, 1, 0, 1, 1, 0, 0, false).
		AddItem(messageInput, 2, 0, 1, 1, 0, 0, true)

	c.app.SetRoot(layout, true)
	return nil
}

func (c *Chat) connectWS() error {
	scheme := "wss"
	if strings.HasPrefix(c.serverAddr, "http://") {
		scheme = "ws"
	}

	host := strings.Replace(strings.Replace(c.serverAddr, "http://", "", 1), "https://", "", 1)
	url := url.URL{Scheme: scheme, Host: host, Path: fmt.Sprintf("/ws/%s/streamer", c.username)}
	log.Printf("Openning socket at %s", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return fmt.Errorf("Failed to connected to websocket: %s", err)
	}
	c.conn = conn

	// send client info so server can verify
	clientInfo := message.ClientInfo{
		Name:   c.username,
		Role:   message.RStreamerChat,
		Secret: GetSecret(CONFIG_PATH),
	}

	msg, _ := message.Wrap(message.TClientInfo, clientInfo)
	payload, _ := json.Marshal(msg)
	err = conn.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		return fmt.Errorf("Failed to connect to server")
	}

	// Verify server's response
	_, resp, err := conn.ReadMessage()
	wrappedMsg, err := message.Unwrap(resp)
	log.Printf("Got a message: %s", wrappedMsg)
	if wrappedMsg.Type == message.TStreamerUnauthorized {
		return fmt.Errorf("Unauthorized connection")
	} else if wrappedMsg.Type != message.TStreamerAuthorized {
		return fmt.Errorf("Expect connect confirmation from server")
	}

	return nil
}

func (c *Chat) addChatMsgs(chatList []message.Chat) {
	if len(chatList) == 0 {
		return
	}
	newChat := ""
	for _, chatObj := range chatList {
		newChat += FormatChat(chatObj.Name, chatObj.Content, chatObj.Color)
	}

	currentChat := c.chatTextView.GetText(false)
	if len(currentChat) > 1 && currentChat[len(currentChat)-1] == '\n' {
		currentChat = currentChat[0 : len(currentChat)-1]
	}

	c.chatTextView.SetText(currentChat + newChat)
}

func (c *Chat) Stop() {
	c.conn.Close()
	c.app.Stop()
}

func FormatChat(name, content, color string) string {
	if len(content) == 0 {
		return ""
	}
	content = strings.TrimPrefix(content, "\n")
	log.Printf("content: %s|", content)
	if content[len(content)-1] != '\n' {
		content += "\n"
	}
	return fmt.Sprintf("[%s]%s[white]: %s", color, name, content)
}
