/*
Define message structs for communication
- Streamer <-> Server
- Server <-> Viewers
*/
package message

import (
	"encoding/json"
	"fmt"
	"time"
)

// *** Generic ***

// Message type
type MType string

const (
	TWrite      MType = "Write"
	TChat       MType = "Chat"
	TClose      MType = "Close"
	TError      MType = "Error"
	TRoomInfo   MType = "RoomInfo"
	TClientInfo MType = "ClientInfo"

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

	// Server will when this message if streamer is verified. Then streamer can proceed to start stream
	TStreamerAuthorized MType = "StreamerAuthorized"

	// If websocket connection is illegal. server send this message to streamer then close connection
	TStreamerUnauthorized MType = "StreamerUnauthorized"
)

type Wrapper struct {
	Type MType
	Data []byte
}

type Winsize struct {
	Rows uint16
	Cols uint16
}

type Chat struct {
	Name    string
	Content string
	Color   string
	Time    string
	Role    string
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
}

// *** Client ***
type CRole string

const (
	RStreamerChat CRole = "StreamerChat" // Chat for streamer
	RStreamer     CRole = "Streamer"     // Send content to server
	RViewer       CRole = "Viewer"       // View content + chat
)

type ClientInfo struct {
	Name   string
	Role   CRole
	Secret string
}

// *** Helper functions ***

func Unwrap(buff []byte) (Wrapper, error) {
	obj := Wrapper{}
	err := json.Unmarshal(buff, &obj)
	return obj, err
}

// Unwrap the wrapper data as well
func Unwrap2(buff []byte) (MType, interface{}, error) {
	msg := Wrapper{}
	err := json.Unmarshal(buff, &msg)
	if err != nil {
		return msg.Type, nil, err
	}

	var msgObj interface{}
	switch msg.Type {
	case TChat:
		msgObj = Chat{}
		err = json.Unmarshal(msg.Data, msgObj)

	case TRequestClientInfo:
		msgObj = ClientInfo{}
		err = json.Unmarshal(msg.Data, msgObj)

	default:
		err = fmt.Errorf("Not implemented")
	}

	return msg.Type, msgObj, err
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
