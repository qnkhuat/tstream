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

type Type string

const (
	TWrite    Type = "Write"
	TChat          = "Chat"
	TClose         = "Close"
	TError         = "Error"
	TRoomInfo      = "RoomInfo"

	// When streamer resize their termianl
	TWinsize = "Winsize"

	// Viewer request server to send winsize
	TRequestWinsize = "RequestWinsize"

	TRequestRoomInfo = "RequestRoomInfo"

	// When user first connect to server
	TStreamerConnect = "StreamerConnect"

	// when user first join the room, he can request for cached message to avoid idle screen
	TRequestCacheMessage = "RequestCacheMessage"
)

type Wrapper struct {
	Type Type
	Data []byte
}

type Winsize struct {
	Rows uint16
	Cols uint16
}

type StreamerConnect struct {
	Title string
}

type RoomInfo struct {
	NViewers    int
	StartedTime time.Time
	Title       string
	StreamerID  string
}

func Unwrap(buff []byte) (Wrapper, error) {
	obj := Wrapper{}
	err := json.Unmarshal(buff, &obj)
	return obj, err
}

func Wrap(msgType Type, msgObject interface{}) (Wrapper, error) {

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
