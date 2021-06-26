package server

import (
	"fmt"
	"github.com/gorilla/mux"
	//"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/room"
	"log"
	"net/http"
	"time"
)

type Server struct {
	rooms  []*room.Room
	addr   string
	server *http.Server
}

func New(addr string) *Server {
	rooms := make([]*room.Room, 0)
	return &Server{
		addr:  addr,
		rooms: rooms,
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("health check")
	fmt.Fprintf(w, "I'm fine, go away: %s\n", time.Now().String())
}

//func handleWebSocket(w http.ResponseWriter, r *http.Request) {
//	log.Println("Connecting")
//	httpUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
//	conn, err := httpUpgrader.Upgrade(w, r, nil)
//	defer conn.Close()
//	if err != nil {
//		log.Panicf("Failed to upgrade to websocket: %s", err)
//	}
//
//	// TODO: turn this into a struct that implement READ/WRITE
//	for { // why this function run with a low frequency?
//		buf := make([]byte, 1024)
//		read, err := pty.f.Read(buf)
//
//		msg := MsgWrapper{
//			Data: buf[:read],
//			Type: "Write",
//		}
//
//		payload, err := json.Marshal(msg)
//		if err != nil {
//			log.Panicf("Failed to Encode msg: ", err)
//		}
//		if err != nil {
//			conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
//			log.Panicf("Unable to read from pty/cmd: %s", err)
//			return
//		}
//		conn.WriteMessage(websocket.BinaryMessage, payload)
//		log.Println("Sent a buffer")
//	}
//}

func (s *Server) Start() {
	router := mux.NewRouter()

	router.HandleFunc("/health", healthHandler)
	//router.HandleFunc("/room/{roomID}/wss", handleWebSocket) // for streamers
	//router.HandleFunc("/room/{roomID}/wsv", handleWebSocket) // for viewers

	s.server = &http.Server{Addr: s.addr, Handler: router}
	log.Printf("Serving at: %s", s.addr)
	if err := s.server.ListenAndServe(); err != nil {
		log.Panicf("Faield to start server: %s", err)
	}
}

func (s *Server) Stop() {
	s.server.Close()
}
