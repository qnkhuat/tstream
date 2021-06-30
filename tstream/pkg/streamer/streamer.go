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
	"time"
)

type Streamer struct {
	pty        *ptyMaster.PtyMaster
	serverAddr string
	id         string
	conn       *websocket.Conn
	Out        chan []byte
	In         chan []byte
}

func New(serverAddr, id string) *Streamer {
	pty := ptyMaster.New()
	out := make(chan []byte, 256) // buffer 256 send requests
	in := make(chan []byte, 256)  // buffer 256 send requests

	return &Streamer{
		pty:        pty,
		serverAddr: serverAddr,
		id:         id,
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
	url := url.URL{Scheme: "ws", Host: s.serverAddr, Path: fmt.Sprintf("/%s/wss", s.id)}
	log.Printf("Openning socket at %s", url.String())
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Printf("Failed to open websocket: %s", err)
		return err
	}
	s.conn = conn

	s.pty.MakeRaw()
	defer s.Stop()

	winSize, _ := ptyMaster.GetWinsize(0)
	s.Winsize(winSize.Rows, winSize.Cols)

	s.pty.SetWinChangeCB(func(ws *ptyDevice.Winsize) {
		s.Winsize(ws.Rows, ws.Cols)
	})

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

	// Periodcally refresh the pty to serve new user
	// Also act as a ping
	go func() {
		ticker := time.NewTicker(cfg.STREAMER_REFRESH_INTERVAL * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Printf("Refresh")
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

	payload, err := message.Wrap(msg)
	if err != nil {
		log.Printf("Failed to wrap message: %s", err)
	}
	s.Out <- payload
	return len(data), nil
}

func (s *Streamer) Winsize(rows, cols uint16) {
	winsizeData, _ := json.Marshal(message.Winsize{
		Rows: rows,
		Cols: cols,
	})

	msg := &message.Wrapper{
		Type: message.TWinsize,
		Data: winsizeData,
	}

	payload, err := message.Wrap(msg)
	if err != nil {
		log.Printf("Failed to wrap message: %s", err)
	}

	s.Out <- payload
}
