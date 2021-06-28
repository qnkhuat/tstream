/*
A room is virtual object that wrap one streamer and multiple viewers togethher
*/
package room

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
)

type RoomStatus int

const (
	Live RoomStatus = iota
	Stopped
)

type Room struct {
	viewers map[string]*websocket.Conn
	roomID  string
	status  RoomStatus
}

func New(roomID string) *Room {
	viewers := make(map[string]*websocket.Conn)
	return &Room{
		roomID:  roomID,
		viewers: viewers,
		status:  Live,
	}
}

func (r *Room) AddViewer(viewerID string, conn *websocket.Conn) error {
	_, ok := r.viewers[viewerID]
	if ok {
		return fmt.Errorf("Viewer %s existed", conn)
	}

	r.viewers[viewerID] = conn
	return nil
}

func (r *Room) RemoveViewer(viewerID string) {
	delete(r.viewers, viewerID)
}

func (r *Room) Broadcast(msg *message.Wrapper) {
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to decode message: %s", err)
		return
	}

	for id, conn := range r.viewers {
		err := conn.WriteMessage(websocket.BinaryMessage, payload)
		log.Println("Sent a buffer")
		if err != nil {
			log.Printf("Failed to board case to %s. Closing connection", id)
			conn.Close()
			r.RemoveViewer(id)
		}
	}
}
