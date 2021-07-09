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
	TWrite    MType = "Write"
	TChat     MType = "Chat"
	TClose    MType = "Close"
	TError    MType = "Error"
	TRoomInfo MType = "RoomInfo"

	// When streamer resize their termianl
	TWinsize MType = "Winsize"

	// Viewer request server to send winsize
	TRequestWinsize MType = "RequestWinsize"

	TRequestRoomInfo MType = "RequestRoomInfo"

	// When user first connect to server
	TStreamerConnect MType = "StreamerConnect"

	// when user first join the room, he can request for cached message to avoid idle screen
	TRequestCacheMessage MType = "RequestCacheMessage"
)

type Wrapper struct {
	Type MType
	Data []byte
}

type Winsize struct {
	Rows uint16
	Cols uint16
}

type StreamerConnect struct {
	Title string
}

// *** Room ***

type RoomStatus string

const (
	RStreaming RoomStatus = "Streaming"

	// When user actively close connection. Detected via closemessage
	RStopped RoomStatus = "Stopped"

	// When don't receive ping for a long time
	RDisconnected RoomStatus = "Disconnected"
)

type RoomInfo struct {
	AccNViewers int // Accumulated nviewers
	NViewers    int
	StartedTime time.Time
	StoppedTime time.Time
	Title       string
	StreamerID  string
	Status      RoomStatus
}

// *** Helper functions ***

func Unwrap(buff []byte) (Wrapper, error) {
	obj := Wrapper{}
	err := json.Unmarshal(buff, &obj)
	return obj, err
}

func Wrap(msgType MType, msgObject interface{}) (Wrapper, error) {

	data, err := json.Marshal(msgObject)
	if err != nil {
		return Wrapper{}, err
	}

	msg := Wrapper{
		Type: msgType,
		Data: data,
	}
	return msg, nil
}
