package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
)

// upgrade an http request to websocket
var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  cfg.SERVER_READ_BUFFER_SIZE,
	WriteBufferSize: cfg.SERVER_WRITE_BBUFFER_SIZE,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var decoder = schema.NewDecoder()
var emptyByteArray []byte

const (
	// Time to wait before force close on connection.
	CLOSE_GRACE_PERIOD = 2 * time.Second
)

// Queries:
// - status - string : Status of Room to query. Leave blank to get any
// - n - int         : Number of rooms to get. Set to -1 to get all
// - skip - string   : Number of rooms to skip. Used for paging
type ListRoomQuery struct {
	Status string `schema:"status"`
	N      int    `schema:"n"`
	Skip   int    `schema:"skip"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Printf("health check")
	fmt.Fprintf(w, "I'm fine: %s\n", time.Now().String())
}

func (s *Server) handleListRooms(w http.ResponseWriter, r *http.Request) {
	var q ListRoomQuery
	err := decoder.Decode(&q, r.URL.Query())
	if err != nil {
		log.Printf("Failed to decode query: %s", err)
		http.Error(w, fmt.Sprintf("%s", err), 400)
		return
	}

	var rooms []message.RoomInfo
	switch q.Status {
	case "Stopped":
		rooms, err = s.db.GetRooms([]message.RoomStatus{message.RStopped}, q.Skip, q.N)
	case "Streaming":
		rooms, err = s.db.GetRooms([]message.RoomStatus{message.RStreaming}, q.Skip, q.N)
	case "":
		rooms, err = s.db.GetRooms([]message.RoomStatus{message.RStreaming, message.RStopped}, q.Skip, q.N) // get all
	default:
		http.Error(w, fmt.Sprintf("%s", "Invalid status"), 400)
		return
	}
	json.NewEncoder(w).Encode(rooms)
}

type AddRoomQuery struct {
	Title      string `schema:"title,required"`
	StreamerID string `schema:"streamerID,required"`
	Version    string `schema:"version,required"`
}

type AddRoomBody struct {
	Secret string `schema:secret,required`
}

// Websocket connetion from streamer
func (s *Server) handleAddRoom(w http.ResponseWriter, r *http.Request) {
	var q AddRoomQuery
	err := decoder.Decode(&q, r.URL.Query())
	if err != nil {
		log.Printf("Failed to decode queries:%s", err)
		return
	}

	// check if version neeeds to be updated
	if compareVer(q.Version, cfg.SERVER_STREAMER_REQUIRED_VERSION) == -1 {
		log.Printf("Streamer version is too old: %s", q.Version)
		http.Error(w, "Upgraded required", 426)
		return
	}

	var b AddRoomBody
	err = json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		log.Printf("Failed to decode body:%s", err)
		http.Error(w, err.Error(), 400)
		return
	}

	if _, ok := s.rooms[q.StreamerID]; !ok {
		s.NewRoom(q.StreamerID, q.Title, b.Secret)
		log.Printf("Added a room %s, %s", q.StreamerID, q.Title)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		if s.rooms[q.StreamerID].Secret() != b.Secret {
			log.Printf("not authorized %s, %s", s.rooms[q.StreamerID].Secret(), b.Secret)
			http.Error(w, "Room existed and you're not authorized to access this room", 401)
			return
		} else {
			log.Printf("Room existed: %s", q.StreamerID)
			http.Error(w, "Room existed", 400)
			return
		}
	}
}

// Websocket connetion from streamer to stream terminal
func (s *Server) handleWSViewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomName := vars["roomName"]
	log.Printf("Client %s entered room: %s", r.RemoteAddr, roomName)
	room, ok := s.rooms[roomName]
	if !ok {
		fmt.Fprintf(w, "Room not existed")
		log.Printf("Room :%s not existed", roomName)
		return
	}
	conn, err := httpUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %s", err)
	}

	viewerID := uuid.New().String()
	room.AddClient(viewerID, message.RViewer, conn)

	// Handle incoming request from user
	room.ReadAndHandleClientMessage(viewerID) // Blocking call
}

// Websocket connection from streamer
// When connected server will wait for a clientinfo message from streamer
// Server then verify this client info's secret.
// If it matches => send back a message type authorized else send back unauthorized
// This has to be happen in the exact order
func (s *Server) handleWSStreamer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomName := vars["roomName"]

	if _, ok := s.rooms[roomName]; !ok {
		http.Error(w, "Room not existed", 400)
		return
	}
	room := s.rooms[roomName]

	conn, err := httpUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %s", err)
	}
	defer conn.Close()

	graceClose := func(message string) {
		log.Printf(message)
		conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
		time.Sleep(CLOSE_GRACE_PERIOD * time.Second)
	}

	// Wait for client response
	_, msg, err := conn.ReadMessage()
	msgObj, err := message.Unwrap(msg)

	if err != nil || msgObj.Type != message.TClientInfo {
		graceClose("Required client info message")
		return
	}

	clientInfo := &message.ClientInfo{}
	err = json.Unmarshal(msgObj.Data, clientInfo)
	if err != nil {
		graceClose("Failed to decode message")
		return
	}

	if clientInfo.Secret != room.Secret() {
		log.Printf("Unauthorized streamer connection")
		sucessMsg, _ := message.Wrap(message.TStreamerUnauthorized, emptyByteArray)
		payload, _ := json.Marshal(sucessMsg)
		conn.WriteMessage(websocket.TextMessage, payload)
		time.Sleep(CLOSE_GRACE_PERIOD * time.Second)
		return
	} else {
		// Connection is authorized

		switch clientInfo.Role {

		case message.RStreamer:
			sucessMsg, _ := message.Wrap(message.TStreamerAuthorized, emptyByteArray)
			payload, _ := json.Marshal(sucessMsg)
			conn.WriteMessage(websocket.TextMessage, payload)

			err = room.AddStreamer(conn)
			if err != nil {
				log.Printf("Failed to add streamer: %s", err)
			}
			room.Start() // Blocking call

		case message.RStreamerChat:
			log.Printf("Got a streamer chat")
			clientID := uuid.New().String()
			room.AddClient(clientID, message.RStreamerChat, conn)

		default:
			log.Printf("Invalid client role: %s", clientInfo.Role)
		}
	}
}

// a > b => 1
// a < b => -1
// a = b => 0
func compareVer(a, b string) int {
	if a == b {
		return 0
	}
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")

	loopMax := len(bs)
	if len(as) > len(bs) {
		loopMax = len(as)
	}
	var ret = 0
	for i := 0; i < loopMax; i++ {
		var x, y string
		if len(as) > i {
			x = as[i]
		}
		if len(bs) > i {
			y = bs[i]
		}
		xi, _ := strconv.Atoi(x)
		yi, _ := strconv.Atoi(y)
		if xi > yi {
			ret = 1
		} else if xi < yi {
			ret = -1
		}
		if ret != 0 {
			break
		}
	}
	return ret
}
