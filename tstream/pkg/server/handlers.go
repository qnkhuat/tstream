package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

const (
	// Time to wait before force close on connection.
	CLOSE_GRACE_PERIOD = 2 * time.Second
)

/*** Health check API ***/
func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Printf("health check")
	fmt.Fprintf(w, "I'm fine: %s\n", time.Now().String())
}

/*** List rooms API ***/
// Queries:
// - status - string : Status of Room to query. Leave blank to get any
// - n - int         : Number of rooms to get. Set to -1 to get all
// - skip - string   : Number of rooms to skip. Used for paging
type ListRoomQuery struct {
	Status  string `schema:"status"`
	N       int    `schema:"n"`
	Skip    int    `schema:"skip"`
	Private bool   `schema:"private"`
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
		rooms, err = s.db.GetRooms([]message.RoomStatus{message.RStopped}, q.Skip, q.N, q.Private)
	case "Streaming":
		rooms, err = s.db.GetRooms([]message.RoomStatus{message.RStreaming}, q.Skip, q.N, q.Private)
	case "":
		rooms, err = s.db.GetRooms([]message.RoomStatus{message.RStreaming, message.RStopped}, q.Skip, q.N, q.Private) // get all
	default:
		http.Error(w, fmt.Sprintf("%s", "Invalid status"), 400)
		return
	}
	json.NewEncoder(w).Encode(rooms)
}

/*** Add room API ***/
type AddRoomQuery struct {
	Title      string `schema:"title,required"`
	StreamerID string `schema:"streamerID,required"`
	Version    string `schema:"version,required"`
	Private    bool   `schema:"private"`
}

type AddRoomBody struct {
	Secret string `schema:"secret,required"`
	Key    string `schema:"key"`
}

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

	if r, ok := s.rooms[q.StreamerID]; !ok {
		if len(b.Secret) == 0 {
			http.Error(w, "Secret must be non-empty", 400)
			return
		}

		if q.Private {
			if len(b.Key) < 6 {
				http.Error(w, "Key must be more than 6 characters", 400)
				return
			}
		}

		_, err := s.NewRoom(q.StreamerID, q.Title, b.Secret, q.Private, b.Key)
		if err != nil {
			log.Printf("Failed to add room: %s", err)
			http.Error(w, "Failed to create room", 400)
			return
		}

		log.Printf("Added a room %s, %s, %v", q.StreamerID, q.Title, q.Private)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		if s.rooms[q.StreamerID].Secret() != b.Secret {
			log.Printf("not authorized %s, %s", s.rooms[q.StreamerID].Secret(), b.Secret)
			http.Error(w, "Room existed and you're not authorized to access this room", 401)
			return
		} else {
			// Reset info
			r.SetTitle(q.Title)
			r.SetPrivate(q.Private)
			r.SetKey(b.Key)
			log.Printf("Room existed: %s", q.StreamerID)
			http.Error(w, "Room existed", 400)
			return
		}
	}
}

/*** Show Room Status ***/
func (s *Server) handleRoomStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomName := vars["roomName"]
	if room, ok := s.rooms[roomName]; ok {
		json.NewEncoder(w).Encode(room.Summary())
		return
	} else {
		http.Error(w, "Room not existed", 400)
		return
	}
}

/*** Websocket connection for
- streamer
- streamerChat
- producerRTC
- viewer
- consumerRTC

 Any connection require the client to send a ClientInfo message.
 Then Server Will use provivded info to handle the webconnection accordingingly

 If connection come from Streamer or producerRTC => then server will verify the client's secret with rooom's secret
 This has to be happen in the exact order
***/
type viewRoom struct {
	Key string `schema:"key"`
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomName := vars["roomName"]

	log.Printf("new connection at room :%s", roomName)
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

	// Wait for client info
	msg := message.Wrapper{}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	err = conn.ReadJSON(&msg)
	conn.SetReadDeadline(time.Time{}) // reset, there will be no time out for future request

	if err != nil || msg.Type != message.TClientInfo {
		graceClose(conn, fmt.Sprintf("Required client info message, got : %s", msg.Type))
		return
	}

	clientInfo := message.ClientInfo{}
	err = message.ToStruct(msg.Data, &clientInfo)
	if err != nil {
		graceClose(conn, "Failed to decode message")
		return
	}

	// response = true to send back a confirmation
	isAuthorized := func(clientSecret, roomSecret string) bool {
		yes := clientSecret == roomSecret

		var payload message.Wrapper
		if yes {
			payload = message.Wrapper{Type: message.TAuthorized, Data: ""}
		} else {
			payload = message.Wrapper{Type: message.TUnauthorized, Data: ""}
		}
		conn.WriteJSON(payload)
		return yes
	}

	// Add WsConn to room based on client role
	switch clientRole := clientInfo.Role; clientRole {

	case message.RStreamer:
		if isAuthorized(clientInfo.Secret, room.Secret()) {
			err = room.AddStreamer(conn)
			if err != nil {
				log.Printf("Failed to add streamer: %s", err)
			}
			room.Start(s.playbackDir) // Blocking call
		} else {
			graceClose(conn, "")
			log.Printf("Unauthorized: %s", clientRole)
		}
		return

	case message.RStreamerChat, message.RProducerRTC:
		if isAuthorized(clientInfo.Secret, room.Secret()) {
			clientID := room.NewClientID()
			room.AddClient(clientID, clientRole, conn) // Blocking call
		} else {
			graceClose(conn, "Unauthorized")
			log.Printf("Unauthorized: %s", clientRole)
		}
		return

	case message.RViewer, message.RConsumerRTC:
		if room.Private() && !isAuthorized(clientInfo.Key, room.Key()) {
			graceClose(conn, "Unauthorized")
			log.Printf("Unauthorized: %s", clientRole)
		} else {
			clientID := room.NewClientID()
			room.AddClient(clientID, clientRole, conn) // Blocking call
		}
		return

	default:
		log.Printf("Invalid client role: %s", clientInfo.Role)
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

func graceClose(conn *websocket.Conn, message string) {
	conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
	time.Sleep(CLOSE_GRACE_PERIOD * time.Second)
}
