package exWebSocket // Extended websocket

import (
	"github.com/gorilla/websocket"
)

type EXWebSocket struct {
	websocket.Conn
}

//func (ws *EXWebSocket) Write(data []byte) (int, error) {
//	err := ws.WriteMessage(websocket.TextMessage, data)
//	return len(data), err
//}

func (ws *EXWebSocket) Write(data []byte) (int, error) {
	err := ws.WriteMessage(websocket.TextMessage, data)
	return len(data), err
}

func New(conn *websocket.Conn) *EXWebSocket {
	return &EXWebSocket{*conn}
}
