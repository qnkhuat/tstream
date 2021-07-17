/*
A room is virtual object that wrap one streamer and multiple viewers togethher
*/
package room

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"sync"
	"time"
)

var emptyByteArray []byte

type Room struct {
	lock           sync.Mutex
	streamer       *websocket.Conn
	clients        map[string]*Client // Chats + viewrer connection
	accViewers     uint64             // accumulated viewers
	name           string             // also is streamerID
	id             uint64             // Id in DB
	title          string
	lastWinsize    message.Winsize
	startedTime    time.Time
	lastActiveTime time.Time
	msgBuffer      []message.Wrapper
	cacheChat      []message.Chat
	status         message.RoomStatus
	secret         string // used to verify streamer
}

func New(name, title, secret string) *Room {
	clients := make(map[string]*Client)
	var buffer []message.Wrapper
	var cacheChat []message.Chat
	return &Room{
		name:           name,
		accViewers:     0,
		clients:        clients,
		lastActiveTime: time.Now(),
		startedTime:    time.Now(),
		msgBuffer:      buffer,
		status:         message.RStreaming,
		title:          title,
		secret:         secret,
		cacheChat:      cacheChat,
	}
}

func (r *Room) LastActiveTime() time.Time {
	return r.lastActiveTime
}

func (r *Room) StartedTime() time.Time {
	return r.startedTime
}

func (r *Room) Clients() map[string]*Client {
	return r.clients
}

func (r *Room) Id() uint64 {
	return r.id
}

func (r *Room) Secret() string {
	return r.secret
}

func (r *Room) NViewers() int {
	count := 0
	for _, client := range r.clients {
		if client.Role() == message.RViewer {
			count += 1
		}
	}
	return count
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
	if r.streamer != nil {
		r.streamer.Close()
	}
	// Verify streamer secret

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

func (r *Room) AddClient(ID string, role message.CRole, conn *websocket.Conn) error {
	_, ok := r.clients[ID]
	if ok {
		return fmt.Errorf("Client %s existed", conn)
	}

	if role == message.RViewer {
		r.accViewers += 1
	} else if role == message.RStreamerChat {
	} else {
		return fmt.Errorf("Invalid client role: %s", role)
	}

	v := NewClient(role, conn)
	r.clients[ID] = v
	go v.Start()

	return nil
}

func (r *Room) RemoveClient(ID string) error {
	_, ok := r.clients[ID]
	if !ok {
		return fmt.Errorf("CLient %s not found", ID)
	}

	r.lock.Lock()
	delete(r.clients, ID)
	r.lock.Unlock()
	return nil
}

// Wait for request from streamer and broadcast those message to clients
func (r *Room) Start() {
	for {
		msg := message.Wrapper{}
		err := r.streamer.ReadJSON(&msg)
		log.Printf("got message: %v", msg)

		if err != nil {
			log.Printf("Failed to receive message from streamer: %s. Closing. Error: %s", r.name, err)
			r.streamer.Close()
			return
		}

		log.Printf("Server got message: %s", msg.Type)

		switch msgType := msg.Type; msgType {
		case message.TWinsize:

			winsize := message.Winsize{}
			err = message.ToStruct(msg.Data, &winsize)

			if err == nil {
				r.lastWinsize = winsize
				r.addMsgBuffer(msg)
				r.lastActiveTime = time.Now()
				r.Broadcast(msg, []message.CRole{message.RViewer}, []string{})
			} else {
				log.Printf("Failed to decode winsize message: %s", err)
			}

		case message.TWrite:

			r.addMsgBuffer(msg)
			r.lastActiveTime = time.Now()
			r.Broadcast(msg, []message.CRole{message.RViewer}, []string{})

		default:
			log.Printf("Unknown message type: %s", msgType)
		}
	}
}

func (r *Room) addMsgBuffer(msg message.Wrapper) {
	if len(r.msgBuffer) > cfg.ROOM_BUFFER_SIZE {
		r.msgBuffer = r.msgBuffer[1:]
	}
	r.msgBuffer = append(r.msgBuffer, msg)
}

func (r *Room) addCacheChat(chat message.Chat) {
	if len(r.cacheChat) >= cfg.ROOM_CACHE_MSG_SIZE {
		r.cacheChat = r.cacheChat[1:]
	}
	r.cacheChat = append(r.cacheChat, chat)
}

func (r *Room) ReadAndHandleClientMessage(ID string) {
	client, ok := r.clients[ID]
	if !ok {
		return
	}
	for {
		msg, _ := <-client.In

		switch msgType := msg.Type; msgType {
		case message.TRequestWinsize:

			payload := message.Wrapper{
				Type: message.TWinsize,
				Data: message.Winsize{
					Rows: r.lastWinsize.Rows,
					Cols: r.lastWinsize.Cols,
				},
			}
			client.Out <- payload

		case message.TRequestCacheContent:
			// Send msg buffer so clients doesn't face a idle screen when first started
			for _, msg := range r.msgBuffer {
				client.Out <- msg
			}

		case message.TRequestRoomInfo:

			roomInfo := r.PrepareRoomInfo()
			payload := message.Wrapper{
				Type: message.TRoomInfo,
				Data: roomInfo,
			}

			client.Out <- payload

		case message.TRequestCacheChat:

			msg, err := message.Wrap(message.TChat, r.cacheChat)
			if err == nil {
				client.Out <- msg
			} else {
				log.Printf("Error wrapping room info message: %s", err)
			}

		case message.TChat:
			var chatList []message.Chat

			err := message.ToStruct(msg.Data, &chatList)

			if err != nil {
				log.Printf("Error: %s", err)
			}

			for _, chat := range chatList {
				r.addCacheChat(chat)
			}

			r.Broadcast(msg, []message.CRole{message.RViewer, message.RStreamerChat}, []string{ID})
		case message.TRoomUpdate:
			if client.Role() != message.RStreamerChat && client.Role() != message.RStreamer {
				log.Printf("Unauthorized set room title")
				continue
			}

			newRoomInfo, ok := msg.Data.(message.RoomInfo)
			if !ok {
				log.Printf("Failed to decode roominfo: %s")
				continue
			} else {
				r.title = newRoomInfo.Title
				roomInfo := r.PrepareRoomInfo()
				payload := message.Wrapper{
					Type: message.TRoomInfo,
					Data: roomInfo,
				}
				// Broadcast to all participants
				r.Broadcast(payload,
					[]message.CRole{message.RStreamer, message.RStreamerChat, message.RViewer},
					[]string{})
			}

		default:
			log.Printf("Unknown message type :%s", msgType)

		}
	}
}

func (r *Room) Broadcast(msg message.Wrapper, roles []message.CRole, IDExclude []string) {

	for id, client := range r.clients {
		// Check if client is in the list of roles to broadcast
		found := false
		for _, role := range roles {
			if role == client.Role() {
				found = true
				break
			}
		}

		if !found {
			continue
		}

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

		if client.Alive() {
			log.Printf("Broadcasting to: %s", id)
			client.Out <- msg
		} else {
			log.Printf("Failed to boardcast to %s. Closing connection", id)
			r.RemoveClient(id)
		}
	}
}

func (r *Room) Stop(status message.RoomStatus) {
	log.Printf("Stopping room: %s, with Status: %s", r.name, status)
	r.status = status
	for id, client := range r.clients {
		client.Close()
		r.RemoveClient(id)
	}
	r.lock.Lock()
	r.streamer.Close()
	r.lock.Unlock()
}

func (r *Room) PrepareRoomInfo() message.RoomInfo {
	return message.RoomInfo{
		Id:             r.id,
		Title:          r.title,
		NViewers:       r.NViewers(),
		StartedTime:    r.startedTime,
		LastActiveTime: r.lastActiveTime,
		StreamerID:     r.name,
		Status:         r.status,
		AccNViewers:    r.accViewers,
	}
}
