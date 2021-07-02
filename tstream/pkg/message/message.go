/*
Define message structs for communication
- Streamer <-> Server
- Server <-> Viewers
*/
package message

import (
	"encoding/json"
)

type Type string

const (
	TWrite   Type = "Write"
	TChat         = "Chat"
	TWinsize      = "Winsize"
	TClose        = "Close"
	TError        = "Error"
)

type Wrapper struct {
	Type Type
	Data []byte
}

type Winsize struct {
	Rows uint16
	Cols uint16
}

type Chat struct {
	Name string
	Content string
}

func Unwrap(buff []byte) (Wrapper, error) {
	obj := Wrapper{}
	err := json.Unmarshal(buff, &obj)
	return obj, err
}

func Wrap(msg *Wrapper) ([]byte, error) {
	return json.Marshal(msg)
}
