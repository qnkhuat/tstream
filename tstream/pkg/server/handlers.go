package server

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Room struct {
	ID             string    `json:"ID"`
	LastActiveTime time.Time `json:"lastActiveTime"`
	StartedTime    time.Time `json:"startedTime"`
	NViewers       int       `json:"nViewers"`
	Title          string    `json:"title"`
}

func (s *Server) handleListRooms(w http.ResponseWriter, r *http.Request) {
	var data []Room
	for _, room := range s.rooms {
		data = append(data, Room{
			ID:             room.ID,
			LastActiveTime: room.LastActiveTime(),
			StartedTime:    room.StartedTime(),
			NViewers:       len(room.Viewers()),
			Title:          "YOOOOOOO",
		})
	}
	json.NewEncoder(w).Encode(data)
}

// Websocket connetion from streamer
func (s *Server) handleWSViewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomID"]
	log.Printf("Client %s entered room: %s", r.RemoteAddr, roomID)
	room, ok := s.rooms[roomID]
	if !ok {
		fmt.Fprintf(w, "Room not existed")
		log.Printf("Room :%s not existed", roomID)
		return
	}
	httpUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := httpUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panicf("Failed to upgrade to websocket: %s", err)
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
	roomID := vars["roomID"]
	if _, ok := s.rooms[roomID]; !ok {
		s.NewRoom(roomID)
	}

	httpUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := httpUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panicf("Failed to upgrade to websocket: %s", err)
	}
	defer conn.Close()
	err = s.rooms[roomID].AddStreamer(conn)
	if err != nil {
		log.Panicf("Failed to add streamer: %s", err)
	}

	s.rooms[roomID].Start() // Blocking call
}
