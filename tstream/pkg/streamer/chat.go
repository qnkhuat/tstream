/* Chat client on terminal */
package streamer

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
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
	c.InitUI()

	if err := c.app.EnableMouse(true).Run(); err != nil {
		panic(err)
	}
	return nil
}

func (c *Chat) InitUI() error {
	layout := tview.NewGrid().
		SetRows(3, 0, 1).
		SetColumns(0).
		SetBorders(true)

	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}
	menu := newPrimitive("Menu")
	ChatTextView := tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true).
		SetWordWrap(true).SetText("a\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\na\nb\nc\nd\ne\nf\ng\nh\n").
		ScrollToEnd()
	//sideBar := newPrimitive("Side Bar")
	messageInput := tview.NewInputField()
	messageInput.SetLabel("[red]>[red] ").
		SetDoneFunc(func(key tcell.Key) {
			messageInput.SetText("")
		})

	layout.AddItem(menu, 0, 0, 1, 1, 0, 0, false).
		AddItem(ChatTextView, 1, 0, 1, 1, 0, 0, false).
		AddItem(messageInput, 2, 0, 1, 1, 0, 0, true)

	c.app.SetRoot(layout, true)
	return nil
}

//func main() {
//	app := tview.NewApplication()
//	inputField := tview.NewInputField().
//		SetLabel("Enter a number: ").
//		SetPlaceholder("E.g. 1234").
//		SetFieldWidth(10).
//		SetAcceptanceFunc(tview.InputFieldInteger).
//		SetDoneFunc(func(key tcell.Key) {
//			app.Stop()
//		})
//	if err := app.SetRoot(inputField, true).EnableMouse(true).Run(); err != nil {
//		panic(err)
//	}
//}
