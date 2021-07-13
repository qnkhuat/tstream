/*
Generic struct for a websocket connection
Currently used for Viewer and Chat
*/
package room

import (
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"time"
)

type Client struct {
	conn *websocket.Conn
	role message.CRole

	// data go in Out channel will be send to user via websocket
	Out chan []byte

	// Data sent from user will be stored in In channel
	In chan []byte

	alive bool
}

func NewClient(role message.CRole, conn *websocket.Conn) *Client {
	out := make(chan []byte, 256) // buffer 256 send requests
	in := make(chan []byte, 256)  // buffer 256 send requests
	return &Client{
		conn:  conn,
		Out:   out,
		In:    in,
		role:  role,
		alive: true,
	}
}

func (v *Client) Role() message.CRole {
	return v.role
}

func (v *Client) Alive() bool {
	return v.alive
}

func (v *Client) Start() {
	// Receive message coroutine
	go func() {
		for {
			msg, ok := <-v.Out
			if ok {
				err := v.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Printf("Failed to boardcast to. Closing connection")
					v.Close()
				}
			} else {
				v.Close()
			}
		}
	}()

	// Send message coroutine
	for {
		_, msg, err := v.conn.ReadMessage()
		if err == nil {
			v.In <- msg // Will be handled in Room
		} else {
			log.Printf("Closing connection")
			v.Close()
			break
		}
	}
}

func (v *Client) Close() {
	v.conn.WriteControl(websocket.CloseMessage, emptyByteArray, time.Time{})
	v.alive = false
	v.conn.Close()
}
