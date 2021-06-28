package server

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/message"
	"github.com/qnkhuat/tstream/pkg/room"
	"log"
	"net/http"
	"time"
)

type Server struct {
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
	fmt.Fprintf(w, "I'm fine, go away: %s\n", time.Now().String())
}

// upgrade an http request to websocket
var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) UpdateWindowSize(rows, cols int) {

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

	s.rooms[roomID].AddViewer(uuid.New().String(), conn)
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
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Failed to read message: %s", err)
			conn.Close()
			return
		}
		msgW := &message.Wrapper{
			Type: message.TWrite,
			Data: msg,
		}
		go s.rooms[roomID].Broadcast(msgW)
	}
}

func (s *Server) Start() {
	router := mux.NewRouter()

	router.HandleFunc("/health", handleHealth)
	router.HandleFunc("/r/{roomID}/wss", s.handleWSServer) // for streamers
	router.HandleFunc("/r/{roomID}/wsv", s.handleWSViewer) // for viewers

	s.server = &http.Server{Addr: s.addr, Handler: router}
	log.Printf("Serving at: %s", s.addr)
	if err := s.server.ListenAndServe(); err != nil {
		log.Panicf("Faield to start server: %s", err)
	}
}

func (s *Server) Stop() {
	s.server.Close()
}
