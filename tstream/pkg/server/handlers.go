package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"net/http"
	"time"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Printf("health check")
	fmt.Fprintf(w, "I'm fine: %s\n", time.Now().String())
}

// upgrade an http request to websocket
var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  cfg.SERVER_READ_BUFFER_SIZE,
	WriteBufferSize: cfg.SERVER_WRITE_BBUFFER_SIZE,
  CheckOrigin: func(r *http.Request) bool { 
    return true 
  },
}

var emptyByteArray []byte
var decoder = schema.NewDecoder()

type Room struct {
	StreamerID     string             `json:"streamerID"`
	LastActiveTime time.Time          `json:"lastActiveTime"`
	StartedTime    time.Time          `json:"startedTime"`
	NViewers       int                `json:"nViewers"`
	Title          string             `json:"title"`
	Status         message.RoomStatus `json:"roomStatus"`
}

// Queries:
// - status - string : Status of Room to query. Leave blank to get any
// - n - int         : Number of rooms to get. Set to -1 to get all
// - skip - string   : Number of rooms to skip. Used for paging
type ListRoomQuery struct {
	Status string `schema:"status"`
	N      int    `schema:"n"`
	Skip   int    `schema:"skip"`
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
	Title      string  `schema:"title,required"`
	StreamerID string  `schema:"streamerId,required"`
  // Override existing room if found one currently streaming
  Force      bool    `schema:"force,required"`
}

// Websocket connetion from streamer
func (s *Server) handleAddRoom(w http.ResponseWriter, r *http.Request) {
	var q AddRoomQuery
	decoder.Decode(&q, r.URL.Query())
  log.Printf("Got query: %v", q)

	if _, ok := s.rooms[q.StreamerID]; !ok || q.Force {
		s.NewRoom(q.StreamerID, q.Title)
		log.Printf("added a room %s, %s", q.Title, q.StreamerID)
	} else {
		log.Printf("Room existed")
		http.Error(w, "Room existed", 400)
	}
  
}

// Websocket connetion from streamer
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

	// Now any message broadcasted to the room will also be broadcast to this connection
	viewerID := uuid.New().String()
	room.AddViewer(viewerID, conn)

	// Handle incoming request from user
	room.ReadAndHandleViewerMessage(viewerID) // Blocking call
}

// Websocket connection from streamer
// TODO: Add key checking to make sure only streamer can stream via this endpoint
func (s *Server) handleWSStreamer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomName := vars["roomName"]

	if _, ok := s.rooms[roomName]; !ok {
		http.Error(w, "Room not existed", 400)
		return
	}

  // TODO config this
	conn, err := httpUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %s", err)
	}
	defer conn.Close()
	err = s.rooms[roomName].AddStreamer(conn)
	if err != nil {
		log.Printf("Failed to add streamer: %s", err)
	}

	s.rooms[roomName].Start() // Blocking call
}
