package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/room"
	"github.com/rs/cors"
)

// TODO: add stream history
// When a room stop, append to that list
// Periodically save to a database
type Server struct {
	lock   sync.RWMutex
	rooms  map[string]*room.Room
	addr   string
	server *http.Server
	db     *DB
}

func New(addr string, db_path string) (*Server, error) {
	rooms := make(map[string]*room.Room)

	db, err := SetupDB(db_path)
	if err != nil {
		log.Printf("Failed to setup database: %s", err)
		return nil, err
	}

	return &Server{
		addr:  addr,
		rooms: rooms,
		db:    db,
	}, nil
}
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set headers
		w.Header().Set("Access-Control-Allow-Headers:", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Next
		next.ServeHTTP(w, r)
		return
	})
}

func (s *Server) NewRoom(name string, title string) error {
	if _, ok := s.rooms[name]; ok {
		return fmt.Errorf("Room %s existed", name)
	}
	s.rooms[name] = room.New(name, title)

	msg := s.rooms[name].PrepareRoomInfo()

	id, err := s.db.PutRoom(msg)
	if err != nil {
		log.Println("Failed to add room to database")
		return err
	}
	s.rooms[name].SetId(id)
	return nil
}

func (s *Server) Start() {
	log.Printf("Serving at: %s", s.addr)
	router := mux.NewRouter()
	router.Use(CORS)

	router.HandleFunc("/api/health", handleHealth).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/rooms", s.handleListRooms).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/room", s.handleAddRoom).Queries("streamerID", "{streamerID}", "title", "{title}").Methods("POST", "OPTIONS")
	router.HandleFunc("/ws/{roomName}/streamer", s.handleWSStreamer) // for streamers
	router.HandleFunc("/ws/{roomName}/viewer", s.handleWSViewer)     // for viewers
	handler := cors.Default().Handler(router)

	s.server = &http.Server{Addr: s.addr, Handler: handler}

	// Scan every SERVER_CLEAN_INTERVAL seconds and delete rooms that idle more than SERVER_CLEAN_THRESHOLD minutes
	go s.cleanRooms(cfg.SERVER_CLEAN_INTERVAL, cfg.SERVER_CLEAN_THRESHOLD)

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
	for roomName, room := range s.rooms {
		if time.Since(room.LastActiveTime()) > threshold || room.Status() == message.RStopped {
			s.deleteRoom(roomName)
			room.SetStatus(message.RStopped) // in case room are current disconnected
			msg := room.PrepareRoomInfo()
			s.db.UpdateRoom(room.Id(), msg)
			count += 1
			log.Printf("Removed room: %s because of Idle", roomName)
		}
	}
	return count
}

func (s *Server) deleteRoom(name string) {
	s.lock.Lock()
	delete(s.rooms, name)
	s.lock.Unlock()

}
