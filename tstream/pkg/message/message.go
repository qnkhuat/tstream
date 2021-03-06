/*
Define message structs for communication
- Streamer <-> Server
- Server <-> Viewers
*/
package message

import (
	"encoding/json"
	"time"
)

// *** Generic ***

// Message type
type MType string

const (
	TWrite      MType = "Write"
	TWriteBlock MType = "WriteBlock"
	TChat       MType = "Chat"
	TClose      MType = "Close"
	TError      MType = "Error"
	TRoomInfo   MType = "RoomInfo"
	TClientInfo MType = "ClientInfo"
	TRoomUpdate MType = "RoomUpdate"
	TRTC        MType = "RTC"

	// When streamer resize their termianl
	TWinsize MType = "Winsize"

	// Viewer request server to send winsize
	TRequestWinsize MType = "RequestWinsize"

	TRequestRoomInfo MType = "RequestRoomInfo"

	// when user first join the room, he can request for cached message to avoid idle screen
	TRequestCacheContent MType = "RequestCacheContent"

	// when user first join the room, he can request for cached chat to avoid idle chat screen
	TRequestCacheChat MType = "RequestCacheChat"

	// Server can request client info to assign roles and verrification
	TRequestClientInfo MType = "RequestClientInfo"

	TAuthorized   MType = "Authorized"
	TUnauthorized MType = "Unauthorized"
)

type Wrapper struct {
	Type MType
	Data interface{}

	// time delay of message to take affect
	// this time is relative with the start time of the parent data block it is sent with
	Delay int64 // milliseconds
}

type Winsize struct {
	Rows uint16
	Cols uint16
}

type TermWriteBlock struct {
	StartTime time.Time

	// how many milliseconds of data this block contains
	Duration int64

	// gzipped array of messages
	Data []byte
}

type Chat struct {
	Name    string
	Content string
	Color   string
	Time    string
	Role    CRole
}

// *** Room ***
type RoomStatus string

const (
	RStreaming RoomStatus = "Streaming"

	// When user actively close connection. Detected via closemessage
	RStopped RoomStatus = "Stopped"
)

type RoomInfo struct {
	Id             uint64 // Id in DB
	AccNViewers    uint64 // Accumulated nviewers
	NViewers       int
	StartedTime    time.Time
	LastActiveTime time.Time
	Title          string
	StreamerID     string
	Status         RoomStatus
	Delay          uint64 // Viewer delay time with streamer ( in milliseconds )
	Private        bool
}

// used for streamer to update room info
type RoomUpdate struct {
	Title string
}

// *** Client ***
// Client Roles
type CRole string

const (
	RStreamerChat CRole = "StreamerChat" // Chat for streamer
	RStreamer     CRole = "Streamer"     // Send content to server
	RViewer       CRole = "Viewer"       // View content + chat
	RConsumerRTC  CRole = "RTCConsumer"  // Consumer only RTC connection : viewer listen to room voice chat
	RProducerRTC  CRole = "RTCProducer"  // Publish of RTC conneciton: streamer publish voice in room
)

type ClientInfo struct {
	Name   string
	Role   CRole
	Secret string // used to verify streamer
	Key    string // used to access private room
}

// ** RTC ***
type RTCEvent string

const (
	RTCOffer     RTCEvent = "Offer"
	RTCAnswer    RTCEvent = "Answer"
	RTCCandidate RTCEvent = "Candidate"
)

type RTC struct {
	Event RTCEvent
	Data  string
}

// *** Helper functions ***

func Unwrap(buff []byte) (Wrapper, error) {
	obj := Wrapper{}
	err := json.Unmarshal(buff, &obj)
	return obj, err
}

func Wrap(msgType MType, data interface{}) Wrapper {
	msg := Wrapper{
		Type: msgType,
		Data: data,
	}
	return msg
}

// convert a map to struct
// data is a map
// v is a reference to a typed variable
func ToStruct(data interface{}, v interface{}) error {
	dataByte, _ := json.Marshal(data)
	err := json.Unmarshal(dataByte, v)
	return err
}
