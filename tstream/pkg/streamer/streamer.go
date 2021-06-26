package streamer

import (
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/ptyMaster"
	"io"
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

func (s *Streamer) Start() {
	s.pty.StartShell()

	go func() { io.Copy(s.pty.F(), os.Stdin) }() // Pipe what user type to terminal session
	io.Copy(os.Stdout, s.pty.F())                // Pipe command response to Pty
}

func (s *Streamer) Stop() {
	s.pty.Stop()
	s.ws.Close()

}
