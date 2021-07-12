package streamer

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	ptyDevice "github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/ptyMaster"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// TODO: if we supports windows this should be changed
var CONFIG_PATH = os.ExpandEnv("$HOME/.tstream.conf")

type Streamer struct {
	pty        *ptyMaster.PtyMaster
	serverAddr string
	clientAddr string
	username   string
	secret     string
	title      string
	conn       *websocket.Conn
	Out        chan []byte
	In         chan []byte
}

func New(clientAddr, serverAddr, username, title string) *Streamer {
	pty := ptyMaster.New()
	out := make(chan []byte, 256) // buffer 256 send requests
	in := make(chan []byte, 256)  // buffer 256 send requests

	secret, _ := GetSecret(CONFIG_PATH)

	return &Streamer{
		secret:     secret,
		pty:        pty,
		serverAddr: serverAddr,
		clientAddr: clientAddr,
		username:   username,
		title:      title,
		Out:        out,
		In:         in,
	}
}

var emptyByteArray []byte
var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  cfg.STREAMER_READ_BUFFER_SIZE,
	WriteBufferSize: cfg.STREAMER_WRITE_BBUFFER_SIZE,
}

func (s *Streamer) Start() error {
	envVars := []string{fmt.Sprintf("%s=%s", cfg.STREAMER_ENVKEY_SESSIONID, s.username)}
	s.pty.StartShell(envVars)
	fmt.Printf("Press Enter to continue!")
	bufio.NewReader(os.Stdin).ReadString('\n')

	err := s.ConnectWS()
	if err != nil {
		log.Println(err)
		s.Stop("Failed to connect to server")
	}

	fmt.Printf("🔥 Streaming at: %s/%s\n", s.clientAddr, s.username)

	s.pty.MakeRaw()

	// Send a winsize message at first
	winSize, _ := ptyMaster.GetWinsize(0)
	s.Winsize(winSize.Rows, winSize.Cols)

	// Send a winsize message when ever terminal change size
	s.pty.SetWinChangeCB(func(ws *ptyDevice.Winsize) {
		s.Winsize(ws.Rows, ws.Cols)
	})

	// Send room title
	msg, err := message.Wrap(message.TStreamerConnect, &message.StreamerConnect{Title: s.title})
	if err == nil {
		payload, _ := json.Marshal(msg)
		s.conn.WriteMessage(websocket.TextMessage, payload)
	} else {
		log.Printf("Failed to wrap connect message: %s", err)
	}

	// Pipe command response to Pty and server
	go func() {
		mw := io.MultiWriter(os.Stdout, s)
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
			err := s.conn.WriteMessage(websocket.TextMessage, msg)
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
	// TODO: secure this, otherwise server can control streamer terminal
	go func() {
		_, _, err := s.conn.ReadMessage()
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
				s.pty.Refresh()
			}
		}
	}()

	s.pty.Wait() // Blocking until user exit
	s.Stop("Bye!")
	return nil
}

func (s *Streamer) RequestAddRoom() int {
	// TODO: handle cases when call add api return existed
	body := map[string]string{"secret": s.secret}
	jsonValue, _ := json.Marshal(body)
	payload := bytes.NewBuffer(jsonValue)
	queries := url.Values{
		"streamerID": {s.username},
		"title":      {strings.TrimSpace(s.title)},
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/room?%s", s.serverAddr, queries.Encode()), "application/json", payload)
	if err != nil {
		log.Printf("%s", err)
	}
	log.Printf("%v", resp)
	return resp.StatusCode
}

func (s *Streamer) ConnectWS() error {
	scheme := "wss"
	if strings.HasPrefix(s.serverAddr, "http://") {
		scheme = "ws"
	}

	host := strings.Replace(strings.Replace(s.serverAddr, "http://", "", 1), "https://", "", 1)
	url := url.URL{Scheme: scheme, Host: host, Path: fmt.Sprintf("/ws/%s/streamer", s.username)}
	log.Printf("Openning socket at %s", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return fmt.Errorf("Failed to connected to websocket: %s", err)
	}

	// Handle server ping
	conn.SetPingHandler(func(appData string) error {
		return s.conn.WriteControl(websocket.PongMessage, emptyByteArray, time.Time{})
	})

	s.conn = conn
	return nil
}

func (s *Streamer) Stop(msg string) {
	s.conn.WriteControl(websocket.CloseMessage, emptyByteArray, time.Time{})
	s.conn.Close()
	s.pty.Stop()
	s.pty.Restore()
	fmt.Println()
	fmt.Println(msg)
}

// Default behavior of Write is to send Write message
func (s *Streamer) Write(data []byte) (int, error) {
	msg := &message.Wrapper{
		Type: message.TWrite,
		Data: data,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to wrap message: %s", err)
	}
	s.Out <- payload
	return len(data), nil
}

func (s *Streamer) Winsize(rows, cols uint16) {
	msg, err := message.Wrap(message.TWinsize, message.Winsize{Rows: rows, Cols: cols})
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to wrap message: %s", err)
	}

	s.Out <- payload
}

func GetSecret(configPath string) (string, error) {
	var secret string
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		secret = GenSecret("tstream")
		f, err := os.Create(configPath)
		defer f.Close()
		if err != nil {
			return "", err
		} else {
			f.WriteString(secret)
		}
	} else {
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			return "", err
		}
		secret = string(data)
	}

	return secret, nil
}

func GenSecret(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}
