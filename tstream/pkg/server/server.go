package server

import (
	//"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	//"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/room"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	lock   sync.RWMutex
	rooms  map[string]*room.Room
	addr   string
	server *http.Server
}

func New(addr string) *Server {
	rooms := make(map[string]*room.Room)
	return &Server{
		addr:  addr,
		rooms: rooms,
	}
}

func (s *Server) NewRoom(roomID string) error {
	if _, ok := s.rooms[roomID]; ok {
		return fmt.Errorf("Room %s existed", roomID)
	}
	s.rooms[roomID] = room.New(roomID)
	log.Printf("Created new Room: %s", roomID)
	return nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Printf("health check")
	fmt.Fprintf(w, "I'm fine: %s\n", time.Now().String())
}

// upgrade an http request to websocket
var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) handleWSViewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomID"]
	log.Printf("Client %s entered room: %s", r.RemoteAddr, roomID)
	if _, ok := s.rooms[roomID]; !ok {
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
	s.rooms[roomID].AddViewer(uuid.New().String(), conn)

	// Handle incoming request from user
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

func (s *Server) handleWSServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomID"]
	if _, ok := s.rooms[roomID]; !ok {
		s.NewRoom(roomID)
	}

	log.Println("Connecting")
	httpUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := httpUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panicf("Failed to upgrade to websocket: %s", err)
	}
	defer conn.Close()

	for {
		msgType, msg, err := conn.ReadMessage()
		log.Printf("Got a message: %d", len(msg))

		if err != nil {
			log.Printf("Failed to read message: %s, len: %d, msgType: %d", err, len(msg), msgType)
			conn.Close()
			return
		}
		s.rooms[roomID].Broadcast(msg)
	}
}

func (s *Server) Start() {
	router := mux.NewRouter()

	router.HandleFunc("/health", handleHealth)
	router.HandleFunc("/r/{roomID}/wss", s.handleWSServer) // for streamers
	router.HandleFunc("/r/{roomID}/wsv", s.handleWSViewer) // for viewers

	s.server = &http.Server{Addr: s.addr, Handler: router}
	log.Printf("Serving at: %s", s.addr)

	go s.cleanRooms(60, 10*60) // Scan every 5 seconds and delete rooms that idle more than 10 minutes

	if err := s.server.ListenAndServe(); err != nil { // blocking call
		log.Panicf("Faield to start server: %s", err)
	}
}

func (s *Server) Stop() {
	s.server.Close()
}

// Scan for rooms that are not active and remove from server
// All unit are in seconds
// interval : scan for every interval time
// ildeThreshold : room with idle time above this threshold will be killed
func (s *Server) cleanRooms(interval, idleThreshold int) {
	tick := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-tick.C:
			c := s.scanAndCleanRooms(idleThreshold)
			log.Printf("Cleaned %d rooms", c)
		}
	}
}

func (s *Server) scanAndCleanRooms(idleThreshold int) int {
	threshold := time.Duration(idleThreshold) * time.Second
	count := 0
	for roomID, room := range s.rooms {
		if time.Since(room.LastActive()) > threshold {
			s.deleteRoom(roomID)
			count += 1
			log.Printf("Removed room: %s because of Idle", roomID)
		}
	}
	return count
}

func (s *Server) deleteRoom(roomID string) {
	s.lock.Lock()
	delete(s.rooms, roomID)
	s.lock.Unlock()

}
