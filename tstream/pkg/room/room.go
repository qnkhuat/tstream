/*
A room is virtual object that wrap one streamer and multiple viewers togethher
*/
package room

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/exwebsocket"
	"log"
	"sync"
)

type RoomStatus int

const (
	Live RoomStatus = iota
	Stopped
)

type Room struct {
	mainRWLock sync.RWMutex
	viewers    map[string]*exwebsocket.Conn
	roomID     string
	status     RoomStatus
}

func New(roomID string) *Room {
	viewers := make(map[string]*exwebsocket.Conn)
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

	exConn := exwebsocket.New(conn)
	r.viewers[viewerID] = exConn
	return nil
}

func (r *Room) RemoveViewer(viewerID string) {
	delete(r.viewers, viewerID)
}

func (r *Room) Broadcast(msg []uint8) {
	for id, conn := range r.viewers {
		// TODO: make this for loop run in parallel
		err := conn.SafeWriteMessage(websocket.BinaryMessage, msg)
		log.Println("Sent a buffer")
		if err != nil {
			log.Printf("Failed to board case to %s. Closing connection", id)
			conn.Close()
			r.RemoveViewer(id)
		}
	}
}
