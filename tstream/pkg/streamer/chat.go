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
	chatTextView     *tview.TextView
	nviewersTextView *tview.TextView
	uptimeTextView   *tview.TextView
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
	go func() {
		for {
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("Failed to read message: %s", err)
				// TODO implement stop
				//c.Stop()
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

				newChat := ""
				for _, chatObj := range chatList {
					newChat += FormatChat(chatObj.Name, chatObj.Content, chatObj.Color)
				}

				currentChat := c.chatTextView.GetText(false)

				c.chatTextView.SetText(currentChat + newChat)
			}
		}

	}()

	reqChatMsg := message.Wrapper{
		Type: message.TRequestCacheChat,
		Data: []byte{},
	}
	payload, _ := json.Marshal(reqChatMsg)

	c.conn.WriteMessage(websocket.TextMessage, payload)

	if err := c.app.EnableMouse(true).Run(); err != nil {
		panic(err)
	}

	return nil
}

func (c *Chat) initUI() error {
	layout := tview.NewGrid().
		SetRows(3, 0, 1).
		SetColumns(0).
		SetBorders(true)

	tstreamText := tview.NewTextView().
		SetText("TStream").
		SetTextAlign(tview.AlignCenter)

	titleText := tview.NewTextView().
		SetText("Title")

	usernameText := tview.NewTextView().
		SetText("Username")

	c.nviewersTextView = tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetText("👤 10")

	c.uptimeTextView = tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetText("00:30:31")

	header := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0, 0, 0).
		AddItem(tstreamText, 0, 0, 1, 3, 0, 0, false).
		AddItem(titleText, 1, 0, 2, 2, 0, 0, false).
		AddItem(usernameText, 2, 0, 1, 1, 0, 0, false).
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
			chat := message.Chat{
				Name:    c.username,
				Color:   c.color,
				Content: messageInput.GetText(),
				Time:    time.Now().String(),
				Role:    "Streamer",
			}
			listChat := []message.Chat{chat}

			//for _, value := range iter_array {
			//	curChat := &message.Chat{}
			//	err := json.Unmarshal(value, curChat)
			//	if err != nil {
			//		log.Printf("There's error when unmarshal cache chat %s", err)
			//		continue
			//	}
			//	listChat = append(listChat, *curChat)
			//}

			//	err := json.Unmarshal(value, curChat)
			msg, _ := message.Wrap(message.TChat, listChat)
			payload, _ := json.Marshal(msg)
			c.conn.WriteMessage(websocket.TextMessage, payload)

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

func FormatChat(name, content, color string) string {
	if content[len(content)-1] != '\n' {
		content += "\n"
	}
	return fmt.Sprintf("[%s]%s[white]: %s", color, name, content)
}
