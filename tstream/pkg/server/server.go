package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/qnkhuat/tstream/internal/cfg"
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

func (s *Server) Start() {
	log.Printf("Serving at: %s", s.addr)
	fmt.Printf("Serving at: %s\n", s.addr)
	router := mux.NewRouter()

	router.HandleFunc("/api/health", handleHealth)
	router.HandleFunc("/api/rooms", s.handleListRooms)
	router.HandleFunc("/{roomID}/wss", s.handleWSStreamer) // for streamers
	router.HandleFunc("/{roomID}/wsv", s.handleWSViewer)   // for viewers

	s.server = &http.Server{Addr: s.addr, Handler: router}

	go s.cleanRooms(cfg.SERVER_CLEAN_INTERVAL, cfg.SERVER_CLEAN_THRESHOLD) // Scan every 5 seconds and delete rooms that idle more than 10 minutes

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
		if time.Since(room.LastActiveTime()) > threshold {
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
