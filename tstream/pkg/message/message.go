/*
Define message structs for communication
- Streamer <-> Server
- Server <-> Viewers
*/
package message

type Type string

const (
	TWrite   Type = "Write"
	TWinsize      = "Winsize"
)

type Wrapper struct {
	Type Type
	Data []byte
}

type Winsize struct {
	Rows uint16
	Cols uint16
}
