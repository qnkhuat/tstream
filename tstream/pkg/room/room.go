/*
A room is virtual object that wrap one streamer and multiple viewers togethher
*/
package room

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/viewer"
	"log"
	"sync"
	"time"
)

type Room struct {
	lock           sync.Mutex
	streamer       *websocket.Conn
	viewers        map[string]*viewer.Viewer
	chats          map[string]*viewer.Viewer
	ID             string
	lastWinsize    *message.Winsize
	startedTime    time.Time
	lastActiveTime time.Time
	msgBuffer      [][]byte
}

func New(ID string) *Room {
	viewers := make(map[string]*viewer.Viewer)
	var buffer [][]byte
	return &Room{
		ID:             ID,
		viewers:        viewers,
		lastActiveTime: time.Now(),
		startedTime:    time.Now(),
		msgBuffer:      buffer,
	}
}

func (r *Room) LastActiveTime() time.Time {
	return r.lastActiveTime
}

func (r *Room) StartedTime() time.Time {
	return r.startedTime
}

func (r *Room) Viewers() map[string]*viewer.Viewer {
	return r.viewers
}

func (r *Room) AddStreamer(conn *websocket.Conn) error {
	// TODO: hanlde case when streamer already existed
	if r.streamer != nil {
		r.streamer.Close()
		//return fmt.Errorf("Streamer existed")
	}
	log.Printf("New streamer")
	r.streamer = conn
	r.streamer.SetPingHandler(func(appData string) error {
		r.lastActiveTime = time.Now()
		return nil
	})

	return nil
}

func (r *Room) AddViewer(ID string, conn *websocket.Conn) error {
	_, ok := r.viewers[ID]
	if ok {
		return fmt.Errorf("Viewer %s existed", conn)
	}

	v := viewer.New(ID, conn)
	r.viewers[ID] = v
	go v.Start()

	// send winsize if existed
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
		v.Out <- payload
	}

	// Send msg buffer so viewers doesn't face a idle screen when first started
	for _, msg := range r.msgBuffer {
		v.Out <- msg
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

// Wait for request from streamer and broadcast those message to viewers
func (r *Room) Start() {
	for {
		_, msg, err := r.streamer.ReadMessage()
		log.Printf("Got a message: %d", len(msg))
		if err != nil {
			log.Printf("Failed to reaceive message from streamer: %s. Closing. Error: %s", r.ID, err)
			r.streamer.Close()
			return
		}
		r.lastActiveTime = time.Now()
		r.addMsgBuffer(msg)
		log.Printf("message in buffer %d, %d", len(r.msgBuffer), cap(r.msgBuffer))
		r.Broadcast(msg)
	}
}

func (r *Room) addMsgBuffer(msg []byte) {
	if len(r.msgBuffer) > cfg.ROOM_BUFFER_SIZE {
		r.msgBuffer = r.msgBuffer[1:]
	}
	r.msgBuffer = append(r.msgBuffer, msg)
}

func (r *Room) ReadAndHandleViewerMessage(ID string) {
	viewer, ok := r.viewers[ID]
	if !ok {
		return
	}
	for {
		msg, _ := <-viewer.In
		log.Printf("Room got message: %d", len(msg))
	}
}

func (r *Room) Broadcast(msg []uint8) {

	msgObj, err := message.Unwrap(msg)
	if err == nil && msgObj.Type == message.TWinsize {
		winsize := &message.Winsize{}
		err := json.Unmarshal(msgObj.Data, winsize)
		if err == nil {
			r.lastWinsize = winsize
		}
	}

	count := 0
	for id, viewer := range r.viewers {
		// TODO: make this for loop run in parallel
		if viewer.Alive() {
			count += 1
			viewer.Out <- msg
		} else {
			log.Printf("Failed to boardcast to %s. Closing connection", id)
			r.RemoveViewer(id)
		}
	}
	log.Printf("Broadcasted to %d viewers", count)
}

func (r *Room) Close() {
	for id, _ := range r.viewers {
		r.RemoveViewer(id)
	}
	r.lock.Lock()
	r.streamer.Close()
	r.lock.Unlock()
}
