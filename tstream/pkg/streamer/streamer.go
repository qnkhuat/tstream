package streamer

import (
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/exWebSocket"
	"github.com/qnkhuat/tstream/pkg/ptyMaster"
	"io"
	"log"
	"net/url"
	"os"
)

type Streamer struct {
	pty       *ptyMaster.PtyMaster
	server    string
	sessionID string
	ws        *websocket.Conn
}

func New(host, sessionID string) *Streamer {
	pty := ptyMaster.New()
	return &Streamer{
		pty:       pty,
		server:    host,
		sessionID: sessionID,
	}
}

var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Streamer) Start() error {
	s.pty.StartShell()
	log.Printf("Sucecssfully started shell")

	// Connect socket to server
	u := url.URL{Scheme: "ws", Host: "0.0.0.0:3000", Path: "/r/qnkhuat/wss"}
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	conn := exWebSocket.New(wsConn)
	mw := io.MultiWriter(conn, os.Stdout)

	go func() { io.Copy(s.pty.F(), os.Stdin) }() // Pipe what user type to terminal session
	io.Copy(mw, s.pty.F())                       // Pipe command response to Pty and server
	return nil
}

func (s *Streamer) Stop() {
	s.pty.Stop()
	s.ws.Close()
}
