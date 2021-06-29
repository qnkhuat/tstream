/*
A room is virtual object that wrap one streamer and multiple viewers togethher
*/
package room

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/exwebsocket"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"sync"
	"time"
)

type RoomStatus int

const (
	Live RoomStatus = iota
	Stopped
)

type Room struct {
	lock        sync.Mutex
	streamer    *exwebsocket.Conn
	viewers     map[string]*exwebsocket.Conn
	roomID      string
	status      RoomStatus
	lastWinsize *message.Winsize
	lastActive  time.Time
}

func New(roomID string) *Room {
	viewers := make(map[string]*exwebsocket.Conn)
	return &Room{
		roomID:     roomID,
		viewers:    viewers,
		lastActive: time.Now(),
		status:     Live,
	}
}

func (r *Room) LastActive() time.Time {
	return r.lastActive
}

func (r *Room) RoomID() string {
	return r.roomID
}

func (r *Room) AddStreamer(conn *websocket.Conn) error {
	//if r.streamer != nil {
	//	return fmt.Errorf("Streamer existed")
	//}
	exConn := exwebsocket.New(conn)
	r.streamer = exConn
	return nil
}

func (r *Room) AddViewer(ID string, conn *websocket.Conn) error {
	_, ok := r.viewers[ID]
	if ok {
		return fmt.Errorf("Viewer %s existed", conn)
	}

	exConn := exwebsocket.New(conn)
	r.viewers[ID] = exConn

	if r.lastWinsize != nil {
		winsizeData, _ := json.Marshal(message.Winsize{
			Rows: r.lastWinsize.Rows,
			Cols: r.lastWinsize.Cols,
		})

		msg := &message.Wrapper{
			Type: message.TWinsize,
			Data: winsizeData,
		}
		payload, _ := message.Wrap(msg)

		exConn.SafeWriteMessage(websocket.TextMessage, payload)
	}
	return nil
}

func (r *Room) RemoveViewer(ID string) error {
	_, ok := r.viewers[ID]
	if !ok {
		return fmt.Errorf("Viewer %s not found", ID)
	}

	r.lock.Lock()
	delete(r.viewers, ID)
	r.lock.Unlock()
	return nil
}

func (r *Room) Broadcast(msg []uint8) {
	r.lastActive = time.Now()

	msgObj, err := message.Unwrap(msg)
	if err == nil && msgObj.Type == message.TWinsize {
		winsize := &message.Winsize{}
		err := json.Unmarshal(msgObj.Data, winsize)
		if err == nil {
			r.lastWinsize = winsize
		}
	}

	for id, conn := range r.viewers {
		// TODO: make this for loop run in parallel
		err := conn.SafeWriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("Failed to boardcast to %s. Closing connection", id)
			conn.Close()
			r.RemoveViewer(id)
		}
	}
}

func (r *Room) Close() {
	for id, _ := range r.viewers {
		r.RemoveViewer(id)
	}
}

func (r *Room) ServeContent() {
	for {
		msgType, msg, err := r.streamer.ReadMessage()
		log.Printf("Got a message: %d", len(msg))

		if err != nil {
			log.Printf("Failed to read message: %s, len: %d, msgType: %d", err, len(msg), msgType)
			r.streamer.Close()
			return
		}
		r.Broadcast(msg)
	}
}

func (r *Room) ReadAndHandleViewerMessage(ID string) {
	conn, ok := r.viewers[ID]
	if !ok {
		return
	}
	for {
		msgType, msg, err := conn.ReadMessage()
		log.Printf("Received a message: type:%d, %s", msgType, msg)

		if err != nil {
			log.Printf("Failed to read message: %s, len: %d, msgType: %d", err, len(msg), msgType)
			conn.Close()
			return
		}
	}
}

//func (r *Room)
