package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/room"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Server struct {
	lock        sync.RWMutex
	rooms       map[string]*room.Room
	addr        string
	server      *http.Server
	db          *DB
	playbackDir string
}

func New(addr, dbPath, playbackDir string) (*Server, error) {
	if _, err := os.Stat(playbackDir); os.IsNotExist(err) {
		if err = os.MkdirAll(playbackDir, 0755); err != nil {
			return nil, err
		}
	}

	rooms := make(map[string]*room.Room)

	db, err := SetupDB(dbPath)
	if err != nil {
		log.Printf("Failed to setup database: %s", err)
		return nil, err
	}

	return &Server{
		addr:        addr,
		rooms:       rooms,
		db:          db,
		playbackDir: playbackDir,
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

func (s *Server) NewRoom(name, title, secret string) error {
	if _, ok := s.rooms[name]; ok {
		return fmt.Errorf("Room %s existed", name)
	}
	r := room.New(name, title, secret)

	msg := r.PrepareRoomInfo()
	id, err := s.db.AddRoom(msg)
	r.SetId(id)
	s.rooms[name] = r
	if err != nil {
		log.Println("Failed to add room to database")
		return err
	}
	return nil
}

func (s *Server) Start() {
	log.Printf("Serving at: %s", s.addr)
	fmt.Printf("Serving at: %s\n", s.addr)
	router := mux.NewRouter()
	router.Use(CORS)

	router.HandleFunc("/api/health", handleHealth).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/rooms", s.handleListRooms).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/room/{roomName}/status", s.handleRoomStatus).Methods("GET", "OPTIONS")
	// Add room
	router.HandleFunc("/api/room", s.handleAddRoom).Queries("streamerID", "{streamerID}", "title", "{title}").Methods("POST", "OPTIONS")
	router.HandleFunc("/ws/{roomName}", s.handleWS) // for streamers
	handler := cors.Default().Handler(router)

	s.server = &http.Server{Addr: s.addr, Handler: handler}

	s.scanAndCleanRooms(cfg.SERVER_CLEAN_THRESHOLD)
	s.syncDB()
	go s.repeatedlyCleanRooms(cfg.SERVER_CLEAN_INTERVAL, cfg.SERVER_CLEAN_THRESHOLD)
	go s.repeatedlySyncDB(cfg.SERVER_SYNCDB_INTERVAL)

	if err := s.server.ListenAndServe(); err != nil { // blocking call
		log.Panicf("Failed to start server: %s", err)
		return
	}
}

func (s *Server) Stop() {
	s.server.Close()
}

// Scan for rooms that are not active and remove from server
// All unit are in seconds
// interval : scan for every interval time
// ildeThreshold : room with idle time above this threshold will be killed
func (s *Server) repeatedlyCleanRooms(interval, idleThreshold time.Duration) {
	for _ = range time.Tick(interval) {
		c := s.scanAndCleanRooms(idleThreshold)
		log.Printf("Auto cleaned %d rooms", c)
	}
}

// Clean in active rooms or stopped one
func (s *Server) scanAndCleanRooms(idleThreshold time.Duration) int {
	count := 0
	for roomName, room := range s.rooms {
		if time.Since(room.LastActiveTime()) > idleThreshold || room.Status() == message.RStopped {
			room.Stop(message.RStopped)
			s.deleteRoom(roomName)
			msg := room.PrepareRoomInfo()
			s.db.UpdateRooms(map[uint64]message.RoomInfo{room.Id(): msg})
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

// Periodically sync server state with DB
func (s *Server) repeatedlySyncDB(interval time.Duration) {
	tick := time.NewTicker(interval)
	for {
		select {
		case <-tick.C:
			s.syncDB()
		}
	}
}

func (s *Server) syncDB() {
	toUpdateRooms := map[uint64]message.RoomInfo{}

	// Update all room in RAM
	for _, room := range s.rooms {
		toUpdateRooms[room.Id()] = room.PrepareRoomInfo()
	}

	// check if all streaming rooms inside DB are actually still streaming
	// there is a case where server suddenly die so the streaming rooms inside DB will turn into zoombie state
	// if found, we update its state to stopped
	dbStreamingRooms, err := s.db.GetRooms([]message.RoomStatus{message.RStreaming}, 0, 0)
	if err != nil {
		log.Printf("failed to get rooms from db")
	}
	for _, streamingRoom := range dbStreamingRooms {
		// this case should rarely happens
		if _, found := toUpdateRooms[streamingRoom.Id]; !found {
			streamingRoom.Status = message.RStopped
			toUpdateRooms[streamingRoom.Id] = streamingRoom
		}
	}

	s.db.UpdateRooms(toUpdateRooms)
}
