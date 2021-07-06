package viewer

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

var emptyByteArray []byte

type Viewer struct {
	conn *websocket.Conn
	id   string

	// data go in Out channel will be send to user via websocket
	Out chan []byte

	// Data sent from user will be stored in In channel
	In chan []byte

	alive bool
}

func New(id string, conn *websocket.Conn) *Viewer {
	out := make(chan []byte, 256) // buffer 256 send requests
	in := make(chan []byte, 256)  // buffer 256 send requests
	return &Viewer{
		conn:  conn,
		id:    id,
		Out:   out,
		In:    in,
		alive: true,
	}
}

func (v *Viewer) Alive() bool {
	return v.alive
}

func (v *Viewer) Start() {
	go func() {
		for {
			msg, ok := <-v.Out
			if ok {
				err := v.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Printf("Failed to boardcast to %s. Closing connection", v.id)
					v.Close()
				}
			} else {
				v.Close()
			}
		}
	}()

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

func (v *Viewer) Close() {
	v.conn.WriteControl(websocket.CloseMessage, emptyByteArray, time.Time{})
	v.alive = false
	v.conn.Close()
}
