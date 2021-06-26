package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/pkg/exWebSocket"
	//"github.com/qnkhuat/tstream/pkg/room"
	"log"
	"net/http"
	"time"
)

type Server struct {
	//rooms  map[string]*room.Room
	addr   string
	server *http.Server
}

func New(addr string) *Server {
	//rooms := make(map[string]*room.Room, 0)
	return &Server{
		addr: addr,
		//rooms: rooms,
	}
}

func (s *Server) NewRoom(roomID string) {

}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Printf("health check")
	fmt.Fprintf(w, "I'm fine, go away: %s\n", time.Now().String())
}

func handleWSViewer(w http.ResponseWriter, r *http.Request) {

}

// upgrade an http request to websocket
var httpUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWSServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Printf("%s", vars)
	log.Println("Connecting")
	httpUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	wsConn, err := httpUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panicf("Failed to upgrade to websocket: %s", err)
	}
	conn := exWebSocket.New(wsConn)
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Failed to read message: %s", err)
			conn.Close()
			return
		}
		log.Printf("Recv: (%d)", len(message))
	}

	// TODO: turn this into a struct that implement READ/WRITE
	//for { // why this function run with a low frequency?
	//	buf := make([]byte, 1024)
	//	read, err := pty.f.Read(buf)

	//	msg := MsgWrapper{
	//		Data: buf[:read],
	//		Type: "Write",
	//	}

	//	payload, err := json.Marshal(msg)
	//	if err != nil {
	//		log.Panicf("Failed to Encode msg: ", err)
	//	}
	//	if err != nil {
	//		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	//		log.Panicf("Unable to read from pty/cmd: %s", err)
	//		return
	//	}
	//	conn.WriteMessage(websocket.BinaryMessage, payload)
	//	log.Println("Sent a buffer")
	//}
}

func (s *Server) Start() {
	router := mux.NewRouter()

	router.HandleFunc("/health", handleHealth)
	router.HandleFunc("/r/{roomID}/wss", handleWSServer) // for streamers
	//router.HandleFunc("/r/{roomID}/wsv", handleWebSocket) // for viewers

	s.server = &http.Server{Addr: s.addr, Handler: router}
	log.Printf("Serving at: %s", s.addr)
	if err := s.server.ListenAndServe(); err != nil {
		log.Panicf("Faield to start server: %s", err)
	}
}

func (s *Server) Stop() {
	s.server.Close()
}
