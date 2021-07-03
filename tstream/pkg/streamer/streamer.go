package streamer

import (
	"bufio"
	"encoding/json"
	"fmt"
	ptyDevice "github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/ptyMaster"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type Streamer struct {
	pty        *ptyMaster.PtyMaster
	serverAddr string
	id         string
	title      string
	conn       *websocket.Conn
	Out        chan []byte
	In         chan []byte
}

func New(serverAddr, id, title string) *Streamer {
	pty := ptyMaster.New()
	out := make(chan []byte, 256) // buffer 256 send requests
	in := make(chan []byte, 256)  // buffer 256 send requests

	return &Streamer{
		pty:        pty,
		serverAddr: serverAddr,
		id:         id,
		title:      title,
		Out:        out,
		In:         in,
	}
}

var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  cfg.STREAMER_READ_BUFFER_SIZE,
	WriteBufferSize: cfg.STREAMER_WRITE_BBUFFER_SIZE,
}

func (s *Streamer) Start() error {
	s.pty.StartShell()
	fmt.Printf("Press Enter to continue!\n")
	bufio.NewReader(os.Stdin).ReadString('\n')

	// Connect socket to server
	scheme := "wss"
	if strings.HasPrefix(s.serverAddr, "http://") {
		scheme = "ws"
	}
	host := strings.Replace(strings.Replace(s.serverAddr, "http://", "", 1), "https://", "", 1)
	url := url.URL{Scheme: scheme, Host: host, Path: fmt.Sprintf("/ws/%s/streamer", s.id)}
	log.Printf("Openning socket at %s", url.String())
	fmt.Printf("Openning socket at %s\n", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Printf("Failed to open websocket: %s", err)
		return err
	}
	s.conn = conn

	s.pty.MakeRaw()
	defer s.Stop()

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
		conn.WriteMessage(websocket.TextMessage, payload)
	} else {
		log.Printf("Failed to wrap connect message: %s", err)
	}

	// Pipe command response to Pty and server
	go func() {
		mw := io.MultiWriter(os.Stdout, s)
		_, err := io.Copy(mw, s.pty.F())
		if err != nil {
			log.Printf("Failed to send pty to mw: %s", err)
			s.Stop()
		}
	}()

	// Pipe what user type to terminal session
	go func() {
		_, err := io.Copy(s.pty.F(), os.Stdin)
		if err != nil {
			log.Printf("Failed to send stin to pty: %s", err)
			s.Stop()
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
				log.Printf("Failed to send message: %s", err)
				s.Stop()
				return
			}
		}
	}()

	// Periodcally send a winsize msg to keep alive
	go func() {
		ticker := time.NewTicker(cfg.STREAMER_REFRESH_INTERVAL * time.Second)
		for {
			select {
			case <-ticker.C:
				var emptyByteArray []byte
				s.conn.WriteControl(websocket.PingMessage, emptyByteArray, time.Time{})
				s.pty.Refresh()
			}
		}
	}()

	s.pty.Wait() // Blocking until user exit
	return nil
}

func (s *Streamer) Stop() {
	s.conn.Close()
	s.pty.Stop()
	s.pty.Restore()
	fmt.Println("Bye!")
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
