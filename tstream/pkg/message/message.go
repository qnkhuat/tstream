/*
Define message structs for communication
- Streamer <-> Server
- Server <-> Viewers
*/
package message

type Type string

const (
	TWrite  Type = "Write"
	TResize      = "Resize"
)

type Wrapper struct {
	Type Type
	Data []byte
}
