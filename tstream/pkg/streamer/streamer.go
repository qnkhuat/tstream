/* Streamer package to stream terminal to server */
package streamer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	ptyDevice "github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/ptyMaster"
)

type Streamer struct {
	pty        *ptyMaster.PtyMaster
	serverAddr string
	clientAddr string
	username   string
	secret     string
	title      string
	conn       *websocket.Conn
	recorder   *Recorder
	Out        chan message.Wrapper
	In         chan message.Wrapper
	// delay of sending message in queue
	delay         time.Duration
	blockDuration time.Duration
	private       bool
	key           string // key to access if room is private
}

func New(clientAddr, serverAddr, username, title string) *Streamer {
	pty := ptyMaster.New()
	out := make(chan message.Wrapper, 256) // buffer 256 send requests
	in := make(chan message.Wrapper, 256)  // buffer 256 send requests

	secret := GetSecret(CONFIG_PATH)

	return &Streamer{
		secret:     secret,
		pty:        pty,
		serverAddr: serverAddr,
		clientAddr: clientAddr,
		username:   username,
		title:      title,
		Out:        out,
		In:         in,
		// TODO: no more hardcoding
		delay:         1500 * time.Millisecond,
		blockDuration: 1 * time.Second, // block size has to smaller than delay
	}
}

var emptyByteArray []byte
var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  cfg.STREAMER_READ_BUFFER_SIZE,
	WriteBufferSize: cfg.STREAMER_WRITE_BBUFFER_SIZE,
}

func (s *Streamer) SetPrivate(private bool) {
	s.private = private
}

func (s *Streamer) SetKey(key string) {
	s.key = key
}

func (s *Streamer) Start() error {
	envVars := []string{fmt.Sprintf("%s=%s", cfg.STREAMER_ENVKEY_SESSIONID, s.username)}
	s.pty.StartShell(envVars)
	fmt.Printf("Press Enter to continue!")
	bufio.NewReader(os.Stdin).ReadString('\n')

	// Init websocket connection
	err := s.ConnectWS()
	if err != nil {
		log.Println(err)
		fmt.Println(err.Error())
		s.Stop(err.Error())
		return err
	}

	// Init transporter
	s.recorder = NewRecorder(s.blockDuration, s.delay, s.Out)
	go s.recorder.Start()

	if s.private {
		fmt.Printf("🔥 Streaming at: %s/%s?key=%s\n", s.clientAddr, s.username, s.key)
	} else {
		fmt.Printf("🔥 Streaming at: %s/%s\n", s.clientAddr, s.username)
	}

	s.pty.MakeRaw()

	// Send a winsize message at first
	winSize, _ := ptyMaster.GetWinsize(0)
	s.writeWinsize(winSize.Rows, winSize.Cols)

	// Send a winsize message when ever terminal change size
	s.pty.SetWinChangeCB(func(ws *ptyDevice.Winsize) {
		s.writeWinsize(ws.Rows, ws.Cols)
	})

	// Pipe command response to Pty and server
	go func() {
		//mw := io.MultiWriter(os.Stdout, s, s.tr)
		mw := io.MultiWriter(os.Stdout, s.recorder)
		_, err := io.Copy(mw, s.pty.F())
		if err != nil {
			log.Printf("Failed to send pty to mw: %s", err)
			s.Stop("Failed to connect pty with server\n")
		}
	}()

	// Pipe what user type to terminal session
	go func() {
		_, err := io.Copy(s.pty.F(), os.Stdin)
		if err != nil {
			log.Printf("Failed to send stdin to pty: %s", err)
			s.Stop("Failed to get user input\n")
		}
	}()

	// Send message to server
	go func() {
		for {
			msg, ok := <-s.Out
			if !ok {
				log.Printf("Error while getting message from Out chan")
				continue
			}
			// TODO: Don't close pty when connection failed
			// This make users lose their work while streaming
			err := s.conn.WriteJSON(msg)
			if err != nil {
				log.Printf("Failed to send message. Streamer closing: %s", err)
				time.Sleep(cfg.STREAMER_RETRY_CONNECT_AFTER * time.Second)
				log.Printf("Reconnecting...")
				err = s.ConnectWS()
				if err != nil {
					log.Printf("Failed to retry connection. Closing connection: %s", err)
					s.Stop("Failed to connect with server! Please try again later\n")
					return
				}
			}
		}
	}()

	// Read and handle message from server
	// Current for ping message only
	go func() {
		msg := message.Wrapper{}
		err := s.conn.ReadJSON(&msg)
		log.Printf("Not implemented response for message: %s", msg.Type)
		if err != nil {
			log.Printf("Failed to receive message from server: %s", err)
		}
	}()

	// Periodcally send a winsize msg to keep alive
	go func() {
		ticker := time.NewTicker(cfg.STREAMER_REFRESH_INTERVAL * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Printf("refresh")
				s.pty.Refresh()
			}
		}
	}()

	s.pty.Wait() // Blocking until user exit
	s.Stop("Bye!")
	return nil
}

func (s *Streamer) RequestAddRoom() int {
	body := map[string]string{"secret": s.secret, "key": s.key}
	jsonValue, _ := json.Marshal(body)
	payload := bytes.NewBuffer(jsonValue)
	queries := url.Values{
		"streamerID": {s.username},
		"title":      {strings.TrimSpace(s.title)},
		"version":    {cfg.STREAMER_VERSION},
		"private":    {strconv.FormatBool(s.private)},
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/room?%s", s.serverAddr, queries.Encode()), "application/json", payload)
	if err != nil {
		return 404
	} else {
		return resp.StatusCode
	}
}

// When connect is initlialized, streamer send a client info to server
// Then wait for a confirmation from server for whether or not this connection
// is authorized
func (s *Streamer) ConnectWS() error {
	scheme := "wss"
	if strings.HasPrefix(s.serverAddr, "http://") {
		scheme = "ws"
	}

	host := strings.Replace(strings.Replace(s.serverAddr, "http://", "", 1), "https://", "", 1)
	url := url.URL{Scheme: scheme, Host: host, Path: fmt.Sprintf("/ws/%s", s.username)}
	log.Printf("Openning socket at %s", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	s.conn = conn
	if err != nil {
		return fmt.Errorf("Failed to connect to server")
	}

	//Handle server ping
	conn.SetPingHandler(func(appData string) error {
		return s.conn.WriteControl(websocket.PongMessage, emptyByteArray, time.Time{})
	})

	// Handle server ping
	conn.SetCloseHandler(func(code int, text string) error {
		s.Stop("Closed connection by server")
		return nil
	})

	// send client info so server can verify
	clientInfo := message.ClientInfo{
		Name:   s.username,
		Role:   message.RStreamer,
		Secret: s.secret,
	}

	payload := message.Wrapper{
		Type: message.TClientInfo,
		Data: clientInfo,
	}
	err = conn.WriteJSON(payload)
	if err != nil {
		return fmt.Errorf("Failed to connect to server")
	}

	// Verify server's response
	msg := message.Wrapper{}
	err = conn.ReadJSON(&msg)
	if msg.Type == message.TUnauthorized {
		return fmt.Errorf("Unauthorized connection")
	} else if msg.Type != message.TAuthorized {
		return fmt.Errorf("Expect connect confirmation from server")
	}
	return nil
}

func (s *Streamer) Stop(msg string) {
	if s.conn != nil {
		s.conn.WriteControl(websocket.CloseMessage, emptyByteArray, time.Time{})
		s.conn.Close()
	}

	if s.pty != nil {
		s.pty.Stop()
		s.pty.Restore()
	}

	fmt.Println()
	fmt.Println(msg)
}

func (s *Streamer) writeWinsize(rows, cols uint16) {
	msg := message.Wrapper{
		Type: message.TWinsize,
		Data: message.Winsize{Rows: rows, Cols: cols},
	}

	s.recorder.WriteMsg(msg)

	// send to server so server can easily save it as last winsize
	s.Out <- msg
}
