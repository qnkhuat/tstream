package room

import (
	"github.com/qnkhuat/tstream/pkg/streamer"
)

type Room struct {
	streamer *streamer.Streamer
	//viewers  []*Viewer
	roomID string
}
