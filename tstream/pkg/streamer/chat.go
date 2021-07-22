/* Chat client on terminal */
package streamer

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/qnkhuat/mediadevices"
	"github.com/qnkhuat/mediadevices/pkg/codec/opus"          // This is required to use opus audio encoder
	_ "github.com/qnkhuat/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/rivo/tview"
	"log"
	"math"
	"net/url"
	"strings"
	"time"
)

var decoder = schema.NewDecoder()

type Chat struct {
	username         string
	sessionId        string
	serverAddr       string
	color            string
	wsConn           *websocket.Conn        // for chat and roominfo
	peerConn         *webrtc.PeerConnection // for voice
	app              *tview.Application
	startedTime      time.Time
	chatTextView     *tview.TextView
	nviewersTextView *tview.TextView
	uptimeTextView   *tview.TextView
	titleTextView    *tview.TextView
	muteBtn          *tview.Button
	mute             bool
}

func NewChat(sessionId, serverAddr, username string) *Chat {
	return &Chat{
		username:   username,
		sessionId:  sessionId,
		serverAddr: serverAddr,
		color:      "red",
		app:        tview.NewApplication(),
		mute:       true,
	}
}

func (c *Chat) Start() error {
	c.initUI()

	if err := c.StartChatService(); err != nil {
		log.Printf("Failed to start chat service : %s", err)
		return err
	}

	if err := c.StartVoiceService(); err != nil {
		log.Printf("Failed to start voice service : %s", err)
		c.addNoti(fmt.Sprintf("Failed to start voice service : %s", err))
	}

	// blocking call
	if err := c.app.EnableMouse(true).Run(); err != nil {
		log.Printf("Error in UI app: %s", err)
		return err
	}

	return nil
}

func (c *Chat) StartChatService() error {
	conn, err := c.connectWS(message.RStreamerChat)
	if err != nil {
		log.Printf("Error: %s", err)
		fmt.Printf("Failed to connect to server\n")
		c.app.Stop()
		return err
	}

	c.wsConn = conn

	go func() {
		for {
			msg := message.Wrapper{}
			err := c.wsConn.ReadJSON(&msg)
			if err != nil {
				log.Printf("Failed to read message: %s", err)
				c.Stop(fmt.Sprintf("Failed to read connect to server"))
				return
			}

			switch msg.Type {
			case message.TChat:
				var chatList []message.Chat
				err := message.ToStruct(msg.Data, &chatList)
				if err != nil {
					log.Printf("Failed to decode chat message: %s", err)
					c.Stop("Failed to decode message from server")
					return
				}
				c.addChatMsgs(chatList)
			case message.TRoomInfo:
				roomInfo := message.RoomInfo{}
				err = message.ToStruct(msg.Data, &roomInfo)
				if err != nil {
					log.Printf("Failed to decode roominfo message: %s", err)
					c.Stop("Failed to decode message from server")
					return
				} else {
					c.startedTime = roomInfo.StartedTime
					c.nviewersTextView.SetText(fmt.Sprintf("%d 👤", roomInfo.NViewers))
					c.titleTextView.SetText(fmt.Sprintf("%s", roomInfo.Title))
				}

			default:
				log.Printf("Not implemented to handle message type: %s", msg.Type)

			}
		}
	}()

	c.requestServer(message.TRequestRoomInfo)
	c.requestServer(message.TRequestCacheChat)
	// send this help text after get cache chat
	go func() {
		time.Sleep(1 * time.Second)
		c.addNoti("[yellow]Type /help to get list of available commands[white]")
	}()

	go func() {
		for _ = range time.Tick(5 * time.Second) {
			c.requestServer(message.TRequestRoomInfo)
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
	return nil
}

func (c *Chat) StartVoiceService() error {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{
			URLs: []string{"stun:stun.l.google.com:19302"}},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}

	// Init microphone
	mediaEngine := webrtc.MediaEngine{}
	opusParams, err := opus.NewParams()
	if err != nil {
		return err
	}
	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithAudioEncoders(&opusParams),
	)
	codecSelector.Populate(&mediaEngine)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
	peerConn, err := api.NewPeerConnection(config)
	if err != nil {
		log.Printf("Failed to start webrtc connection %s", err)
		return err
	}

	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {},
		Codec: codecSelector,
	})

	if err != nil {
		log.Printf("Failed to get user media %s", err)
		return err
	}

	// add tracks to peer connection
	for _, track := range s.GetTracks() {
		log.Printf("%s, %s, %s", track.ID(), track.Kind(), track.Kind)
		log.Printf("adding track")
		// TODO: we probably want to stop the chat here
		// reproduce steps: try open a producer page while having chat on
		track.OnEnded(func(err error) {
			log.Printf("Track (ID: %s) ended with error: %v\n",
				track.ID(), err)
		})

		_, err = peerConn.AddTransceiverFromTrack(track,
			webrtc.RtpTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly})
		if err != nil {
			log.Printf("Failed to add track %s", err)
			return err
		}
	}

	// wsconnection is for signaling and update track changes
	wsConn, err := c.connectWS(message.RProducerRTC)
	if err != nil {
		log.Printf("Failed to start voice ws: %s", err)
		return err
	}

	peerConn.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		log.Printf("Peer conenction state change to: %s", p)
		switch p {

		case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed, webrtc.PeerConnectionStateDisconnected:
			c.Stop(fmt.Sprintf("Voice server status: %s", p))

		case webrtc.PeerConnectionStateConnected:

		default:
			log.Printf("Not implemented: %s", p)
		}

	})

	// Trickle ICE. Emit server candidate to client
	peerConn.OnICECandidate(func(ice *webrtc.ICECandidate) {
		if ice == nil {
			return
		}

		candidate, err := json.Marshal(ice.ToJSON())
		if err != nil {
			log.Printf("Failed to decode ice candidate: %s", err)
			return
		}

		payload := message.Wrapper{
			Type: message.TRTC,
			Data: message.RTC{
				Event: message.RTCCandidate,
				Data:  string(candidate),
			}}

		wsConn.WriteJSON(payload)
	})

	// Negotiation
	// Server are going to send the offer first
	go func() {
		for {
			msg := message.Wrapper{}
			err := wsConn.ReadJSON(&msg)
			if err != nil {
				log.Printf("Failed to read message: %s", err)
				c.Stop("Failed to read message form server")
				return
			}

			if msg.Type != message.TRTC {
				log.Printf("Expected RTC Event message, Got :%s", msg.Type)
				continue
			}

			event := message.RTC{}
			if err = message.ToStruct(msg.Data, &event); err != nil {
				log.Printf("Failed to decode RTCevent message")
				continue
			}

			switch eventType := event.Event; eventType {

			case message.RTCOffer:
				// set offer SDP as remote description
				offer := webrtc.SessionDescription{}
				if err := json.Unmarshal([]byte(event.Data), &offer); err != nil {
					log.Println(err)
					return
				}

				if err := peerConn.SetRemoteDescription(offer); err != nil {
					log.Printf("Failed to set remote description: %s", err)
					return
				}

				// send back SDP answer and set it as local description
				answer, err := peerConn.CreateAnswer(nil)
				if err != nil {
					log.Printf("Failed to create Offer")
					return
				}

				if err := peerConn.SetLocalDescription(answer); err != nil {
					log.Printf("Failed to set local description: %v", err)
					return
				}

				answerByte, _ := json.Marshal(answer)
				payload := message.Wrapper{
					Type: message.TRTC,
					Data: message.RTC{
						Event: message.RTCAnswer,
						Data:  string(answerByte),
					},
				}
				if err = wsConn.WriteJSON(payload); err != nil {
					log.Printf("Failed to send answer: %s", err)
				}

			case message.RTCCandidate:
				candidate := webrtc.ICECandidateInit{}
				if err := json.Unmarshal([]byte(event.Data), &candidate); err != nil {
					log.Println(err)
					return
				}

				if err := peerConn.AddICECandidate(candidate); err != nil {
					log.Println(err)
					return
				}

			default:
				log.Printf("Not implemented to handle message type: %s", msg.Type)

			}
		}
	}()
	return nil

}

func (c *Chat) requestServer(msgType message.MType) error {
	payload := message.Wrapper{
		Type: msgType,
		Data: "",
	}
	return c.wsConn.WriteJSON(payload)
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
			text := strings.TrimSpace(messageInput.GetText())
			if len(text) > 0 && text[0] == '/' {
				command := strings.TrimSpace(text[1:])
				c.HandleCommand(command)
				messageInput.SetText("")
				return
			} else {
				chat := message.Chat{
					Name:    c.username,
					Color:   c.color,
					Content: text,
					Time:    time.Now().String(),
					Role:    message.RStreamer,
				}

				chatList := []message.Chat{chat}
				payload := message.Wrapper{Type: message.TChat, Data: chatList}
				c.wsConn.WriteJSON(payload)
				c.addChatMsgs(chatList)
				messageInput.SetText("")
			}
		})

	// Default is mute
	c.muteBtn = tview.NewButton("🔇").
		SetSelectedFunc(func() {
			c.toggleMute()
		})
	c.muteBtn.SetBackgroundColor(tcell.ColorBlack)

	footer := tview.NewGrid().
		SetRows(1).
		SetColumns(3, 0).
		AddItem(c.muteBtn, 0, 0, 1, 1, 0, 0, false).
		AddItem(messageInput, 0, 1, 1, 1, 0, 0, true)

	layout.AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(c.chatTextView, 1, 0, 1, 1, 0, 0, false).
		AddItem(footer, 2, 0, 1, 1, 0, 0, true)

	// auto focus on the message input if user starts typing
	c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			c.app.SetFocus(footer)
		}
		return event

	})
	c.app.SetRoot(layout, true)
	return nil
}

func (c *Chat) HandleCommand(command string) error {
	args := strings.Split(command, " ")
	switch args[0] {
	case "help":
		c.addNoti(`
      TStream - Streaming from terimnal
      [green]/title[yellow] title[white] - to change stream title 
      [green]/mute[white] - to turn on microphone
      [green]/unmute[white] - to turn off microphone
      [green]/exit[white] - to exit chat room`)

	case "title":
		if len(args) > 1 {
			newTitle := strings.Trim(strings.Join(args[1:], " "), "\"")
			roomUpdate := message.RoomUpdate{
				Title: newTitle,
			}
			payload := message.Wrapper{Type: message.TRoomUpdate, Data: roomUpdate}
			err := c.wsConn.WriteJSON(payload)
			if err != nil {
				log.Printf("Failed to set new title : %s", err)
				c.addNoti(`[red]Failed to change title. Please try again[white]`)
			} else {
				c.addNoti(fmt.Sprintf(`[yellow]Changed room title to: %s[white]`, newTitle))
			}
		} else {
			c.addNoti(`[yellow]/title : no title found[white]`)
		}

	case "mute":
		if !c.mute {
			c.toggleMute()
		}
	case "unmute":
		if c.mute {
			c.toggleMute()
		}

	case "exit":
		c.Stop("Bye!")

	default:
		c.addNoti(`Unknown command. Type /help to get list of available commands.`)
	}

	return nil
}

func (c *Chat) ConnctWSVoice() error {
	return nil
}

func (c *Chat) connectWS(role message.CRole) (*websocket.Conn, error) {
	url := getWSUrl(c.serverAddr, c.username)

	log.Printf("Openning socket at %s", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return conn, fmt.Errorf("Failed to connected to websocket: %s", err)
	}

	// send client info so server can verify
	clientInfo := message.ClientInfo{
		Name:   c.username,
		Role:   role,
		Secret: GetSecret(CONFIG_PATH),
	}
	log.Printf("clientinfo %s", clientInfo)

	payload := message.Wrapper{Type: message.TClientInfo, Data: clientInfo}
	err = conn.WriteJSON(payload)
	if err != nil {
		return conn, fmt.Errorf("Failed to connect to server")
	}

	// Verify server's response
	msg := message.Wrapper{}
	err = conn.ReadJSON(&msg)
	if err != nil {
		log.Printf("Failed to read websocket message: %s", err)
		return conn, fmt.Errorf("Failed to read websocket message: %s", err)
	}

	if msg.Type == message.TStreamerUnauthorized {
		return conn, fmt.Errorf("Unauthorized connection")
	} else if msg.Type != message.TStreamerAuthorized {
		return conn, fmt.Errorf("Expect connect confirmation from server")
	}

	return conn, nil
}

func (c *Chat) toggleMute() {
	c.mute = !c.mute
	if c.mute {
		c.muteBtn.SetLabel("🔇")
		c.addNoti(`[yellow]Microphone: On[white]`)
	} else {
		c.muteBtn.SetLabel("🔈")
		c.addNoti(`[yellow]Microphone: Off[white]`)
	}

}

func (c *Chat) addNoti(msg string) {

	if len(msg) > 0 && msg[len(msg)-1] != '\n' {
		msg += "\n"
	}

	currentChat := c.chatTextView.GetText(false)
	if len(currentChat) > 1 && currentChat[len(currentChat)-1] == '\n' {
		currentChat = currentChat[0 : len(currentChat)-1]
	}

	c.chatTextView.SetText(currentChat + msg)
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

func (c *Chat) Stop(msg string) {
	fmt.Printf(msg)
	if c.wsConn != nil {
		c.wsConn.Close()
	}
	if c.peerConn != nil {
		c.peerConn.Close()
	}
	c.app.Stop()
}

func FormatChat(name, content, color string) string {
	if len(content) == 0 {
		return ""
	}
	content = strings.TrimPrefix(content, "\n")
	if content[len(content)-1] != '\n' {
		content += "\n"
	}
	return fmt.Sprintf("[%s]%s[white]: %s", color, name, content)
}

func getWSUrl(serverAddr, username string) string {
	scheme := "wss"
	if strings.HasPrefix(serverAddr, "http://") {
		scheme = "ws"
	}

	host := strings.Replace(strings.Replace(serverAddr, "http://", "", 1), "https://", "", 1)
	url := url.URL{Scheme: scheme, Host: host, Path: fmt.Sprintf("/ws/%s", username)}
	return url.String()
}
