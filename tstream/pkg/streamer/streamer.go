package streamer

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/exWebSocket"
	"github.com/qnkhuat/tstream/pkg/ptyMaster"
	"io"
	//"log"
	"net/url"
	"os"
)

type Streamer struct {
	pty       *ptyMaster.PtyMaster
	server    string
	sessionID string
	ws        *exWebSocket.EXWebSocket
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
	fmt.Printf("Press Enter to continue!\n")
	bufio.NewReader(os.Stdin).ReadString('\n')

	// Connect socket to server
	u := url.URL{Scheme: "ws", Host: "0.0.0.0:3000", Path: "/r/qnkhuat/wss"}
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	conn := exWebSocket.New(wsConn)
	s.ws = conn

	s.pty.MakeRaw()
	defer s.Stop()

	go func() {
		// Pipe command response to Pty and server
		mw := io.MultiWriter(os.Stdout, conn)
		_, err := io.Copy(mw, s.pty.F())
		if err != nil {
			s.Stop()
		}
	}()

	go func() {
		// Pipe what user type to terminal session
		_, err := io.Copy(s.pty.F(), os.Stdin)
		if err != nil {
			s.Stop()
		}
	}()

	s.pty.Wait() // Blocking until user exit
	return nil
}

func (s *Streamer) Stop() {
	s.ws.Close()
	s.pty.Stop()
	s.pty.Restore()
	fmt.Println("Bye!")
}
