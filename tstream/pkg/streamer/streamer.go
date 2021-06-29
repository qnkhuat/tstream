package streamer

import (
	"bufio"
	"fmt"
	ptyDevice "github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/ptyMaster"
	"io"
	"log"
	"net/url"
	"os"
)

type Streamer struct {
	pty        *ptyMaster.PtyMaster
	serverAddr string
	sessionID  string
	sess       *Session
}

func New(serverAddr, sessionID string) *Streamer {
	pty := ptyMaster.New()
	return &Streamer{
		pty:        pty,
		serverAddr: serverAddr,
		sessionID:  sessionID,
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
	url := url.URL{Scheme: "ws", Host: s.serverAddr, Path: fmt.Sprintf("/%s/wss", s.sessionID)}
	log.Printf("Openning socket at %s", url.String())
	wsConn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Printf("Failed to open websocket: %s", err)
		return err
	}
	session := NewSession(wsConn)
	s.sess = session

	s.pty.MakeRaw()
	defer s.Stop()

	winSize, _ := ptyMaster.GetWinsize(0)
	session.Winsize(winSize.Rows, winSize.Cols)

	s.pty.SetWinChangeCB(func(ws *ptyDevice.Winsize) {
		session.Winsize(ws.Rows, ws.Cols)
	})

	go func() {
		// Pipe command response to Pty and server
		mw := io.MultiWriter(os.Stdout, session)
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
	s.sess.Close()
	s.pty.Stop()
	s.pty.Restore()
	fmt.Println("Bye!")
}
