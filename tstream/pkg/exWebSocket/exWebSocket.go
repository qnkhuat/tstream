package exwebsocket

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Conn struct {
	*websocket.Conn
	mu sync.Mutex
}

func New(conn *websocket.Conn) *Conn {
	return &Conn{Conn: conn}
}

func (ws *Conn) SafeWriteMessage(msgType int, data []byte) error {
	ws.mu.Lock()
	err := ws.WriteMessage(websocket.BinaryMessage, data)
	ws.mu.Unlock()
	return err
}
