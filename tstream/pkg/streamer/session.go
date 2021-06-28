package streamer

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/exwebsocket"
	"github.com/qnkhuat/tstream/pkg/message"
)

type Session struct {
	*exwebsocket.Conn
}

func NewSession(conn *websocket.Conn) *Session {
	exConn := exwebsocket.New(conn)
	return &Session{exConn}
}

func (ss *Session) sendMsg(msg *message.Wrapper) error {
	payload, err := message.Wrap(msg)
	if err != nil {
		return err
	}

	err = ss.SafeWriteMessage(websocket.TextMessage, payload)
	return err
}

func (ss *Session) Winsize(rows, cols uint16) error {
	winsizeData, _ := json.Marshal(message.Winsize{
		Rows: rows,
		Cols: cols,
	})

	msg := &message.Wrapper{
		Type: message.TWinsize,
		Data: winsizeData,
	}

	return ss.sendMsg(msg)
}

// Default behavior of Write is to send Write message
func (ss *Session) Write(data []byte) (int, error) {
	msg := &message.Wrapper{
		Type: message.TWrite,
		Data: data,
	}

	err := ss.sendMsg(msg)
	return len(data), err
}
