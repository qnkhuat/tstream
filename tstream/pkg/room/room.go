package room

import (
	"github.com/qnkhuat/tstream/pkg/streamer"
)

type RoomStatus int

const (
	Live RoomStatus = iota
	Stopped
)

type Room struct {
	streamer *streamer.Streamer
	//viewers  []*Viewer
	roomID string
	status RoomStatus
}

func New(roomID string) *Room {
	st := streamer.New(roomID)
	return &Room{
		streamer: st,
	}

}
