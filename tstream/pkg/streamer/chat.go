/* Chat client on terminal */
package streamer

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
	"log"
	"net/url"
	"strings"
)

type Chat struct {
	username   string
	sessionId  string
	serverAddr string
	color      string
	conn       *websocket.Conn
	app        *tview.Application
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

	if err := c.app.EnableMouse(true).Run(); err != nil {
		panic(err)
	}

	c.connectWS()
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

	nviewersText := tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetText("👤 10")

	uptimeText := tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetText("00:30:31")

	header := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0, 0, 0).
		AddItem(tstreamText, 0, 0, 1, 3, 0, 0, false).
		AddItem(titleText, 1, 0, 2, 2, 0, 0, false).
		AddItem(usernameText, 2, 0, 1, 1, 0, 0, false).
		AddItem(nviewersText, 1, 2, 1, 1, 0, 0, false).
		AddItem(uptimeText, 2, 2, 1, 1, 0, 0, false)

	ChatTextView := tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true).
		SetWordWrap(true).SetText("a\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\n").
		ScrollToEnd()

	messageInput := tview.NewInputField()
	messageInput.SetLabel("[red]>[red] ").
		SetDoneFunc(func(key tcell.Key) {
			messageInput.SetText("")
		})

	layout.AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(ChatTextView, 1, 0, 1, 1, 0, 0, false).
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
	url := url.URL{Scheme: scheme, Host: host, Path: fmt.Sprintf("/ws/%s/chat", c.username)}
	log.Printf("Openning socket at %s", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return fmt.Errorf("Failed to connected to websocket: %s", err)
	}
	c.conn = conn
	return nil
}
