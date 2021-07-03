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
	TWrite          Type = "Write"
	TWinsize             = "Winsize"
	TClose               = "Close"
	TError               = "Error"
	TRequestWinsize      = "RequestWinsize"
	TChat				 = "Chat"
)

type Wrapper struct {
	Type Type
	Data []byte
}

type Winsize struct {
	Rows uint16
	Cols uint16
}

func Unwrap(buff []byte) (Wrapper, error) {
	obj := Wrapper{}
	err := json.Unmarshal(buff, &obj)
	return obj, err
}

func Wrap(msg *Wrapper) ([]byte, error) {
	return json.Marshal(msg)
}
