/*
A room is virtual object that wrap one streamer and multiple viewers togethher
*/
package room

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/viewer"
)

var emptyByteArray []byte

type Room struct {
	lock           sync.Mutex
	streamer       *websocket.Conn
	viewers        map[string]*viewer.Viewer
	chats          map[string]*viewer.Viewer
	name           string // also is streamerID
	id             uint64 // Id in DB
	title          string
	lastWinsize    *message.Winsize
	startedTime    time.Time
	stoppedTime    time.Time
	lastActiveTime time.Time
	msgBuffer      [][]byte
	status         message.RoomStatus
}

func New(name string, title string) *Room {
	viewers := make(map[string]*viewer.Viewer)
	var buffer [][]byte
	return &Room{
		name:           name,
		viewers:        viewers,
		lastActiveTime: time.Now(),
		startedTime:    time.Now(),
		msgBuffer:      buffer,
		status:         message.RStreaming,
		title:          title,
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

func (r *Room) Id() uint64 {
	return r.id
}

func (r *Room) SetTitle(title string) {
	r.title = title
}

func (r *Room) SetId(id uint64) {
	r.id = id
}

func (r *Room) SetStatus(status message.RoomStatus) {
	r.status = status
}

func (r *Room) Status() message.RoomStatus {
	return r.status
}

func (r *Room) Title() string {
	return r.title
}

func (r *Room) Streamer() *websocket.Conn {
	return r.streamer
}

func (r *Room) AddStreamer(conn *websocket.Conn) error {
	// TODO: hanlde case when streamer already existed
	if r.streamer != nil {
		r.streamer.Close()
		//return fmt.Errorf("Streamer existed")
	}
	log.Printf("New streamer")
	r.streamer = conn
	r.status = message.RStreaming

	conn.SetPongHandler(func(appData string) error {
		r.lastActiveTime = time.Now()
		return nil
	})

	r.streamer.SetCloseHandler(func(code int, text string) error {
		log.Printf("Got streamer close message. Stopping room: %s", r.name)
		r.status = message.RStopped
		r.Stop(message.RStopped)
		return nil
	})

	// Periodically ping streamer
	// If streamer response with a pong message => still alive
	go func() {
		ticker := time.NewTicker(cfg.SERVER_PING_INTERVAL * time.Second)
		for {
			select {
			case <-ticker.C:
				if r.status == message.RStopped {
					return
				}
				if time.Now().Sub(r.lastActiveTime) > time.Second*cfg.SERVER_DISCONNECTED_THRESHHOLD {
					r.status = message.RStopped
				} else {
					r.status = message.RStreaming
				}
				r.streamer.WriteControl(websocket.PingMessage, emptyByteArray, time.Time{})
			}
		}
	}()

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
		msg, err := message.Wrap(message.TWinsize, message.Winsize{
			Rows: r.lastWinsize.Rows,
			Cols: r.lastWinsize.Cols,
		})

		if err != nil {
			log.Printf("Failed to decode message: %s", err)
		} else {
			payload, _ := json.Marshal(msg)
			v.Out <- payload
		}
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

		if err != nil {
			log.Printf("Failed to reaceive message from streamer: %s. Closing. Error: %s", r.name, err)
			r.streamer.Close()
			return
		}
		wrapperMsg, err := message.Unwrap(msg)
		if err != nil {
			log.Printf("Unable to decode message: %s", err)
			continue
		}
		if wrapperMsg.Type == message.TWinsize || wrapperMsg.Type == message.TWrite {
			r.lastActiveTime = time.Now()
			r.addMsgBuffer(msg)
			r.Broadcast(msg, []string{})
		} else if wrapperMsg.Type == message.TStreamerConnect {
			msgObject := &message.StreamerConnect{}
			err := json.Unmarshal(wrapperMsg.Data, msgObject)
			if err != nil {
				log.Printf("Failed to decode message: %s", err)
			} else {
				r.SetTitle(msgObject.Title)
			}
			log.Printf("Set title: %s", msgObject.Title)

		} else {
			log.Printf("Unknown message type: %s", wrapperMsg.Type)
		}
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

		msgObj, err := message.Unwrap(msg)
		if err != nil {
			log.Printf("Failed to decode msg", err)
		}

		if msgObj.Type == message.TRequestWinsize {

			msg, _ := message.Wrap(message.TWinsize, message.Winsize{
				Rows: r.lastWinsize.Rows,
				Cols: r.lastWinsize.Cols,
			})
			payload, _ := json.Marshal(msg)
			viewer.Out <- payload

		} else if msgObj.Type == message.TRequestCacheMessage {
			// Send msg buffer so viewers doesn't face a idle screen when first started
			for _, msg := range r.msgBuffer {
				viewer.Out <- msg
			}
		} else if msgObj.Type == message.TRequestRoomInfo {

			roomInfo := r.PrepareRoomInfo()
			msg, err := message.Wrap(message.TRoomInfo, roomInfo)

			if err == nil {
				payload, _ := json.Marshal(msg)
				viewer.Out <- payload
			} else {
				log.Printf("Error wrapping room info message: %s", err)
			}
		} else if msgObj.Type == message.TChat {
			r.Broadcast(msg, []string{ID})
		}
	}
}

func (r *Room) Broadcast(msg []uint8, IDExclude []string) {

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
		var isExcluded bool = false
		for _, idExclude := range IDExclude {
			if id == idExclude {
				isExcluded = true
			}
		}
		if isExcluded {
			continue
		}

		if viewer.Alive() {
			count += 1
			viewer.Out <- msg
		} else {
			log.Printf("Failed to boardcast to %s. Closing connection", id)
			r.RemoveViewer(id)
		}
	}
}

func (r *Room) Stop(status message.RoomStatus) {
	log.Printf("Stopping room: %s, with Status: %s", r.name, status)
	r.status = status
	r.stoppedTime = time.Now()
	for id, viewer := range r.viewers {
		viewer.Close()
		r.RemoveViewer(id)
	}
	r.lock.Lock()
	r.streamer.Close()
	r.lock.Unlock()
}

func (r *Room) PrepareRoomInfo() message.RoomInfo {
	return message.RoomInfo{
    Id:          r.Id(),
		Title:       r.title,
		NViewers:    len(r.viewers),
		StartedTime: r.startedTime,
		StreamerID:  r.name,
		Status:      r.status,
	}
}
