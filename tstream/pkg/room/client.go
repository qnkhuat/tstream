/*
Generic struct for a websocket connection
Currently used for Viewer and Chat
*/
package room

import (
	"github.com/gorilla/websocket"
	"github.com/qnkhuat/tstream/internal/cfg"
	"github.com/qnkhuat/tstream/pkg/message"
	"log"
	"time"
)

type Client struct {
	conn *websocket.Conn
	role message.CRole

	// data go in Out channel will be send to user via websocket
	Out chan message.Wrapper

	// Data sent from user will be stored in In channel
	In chan message.Wrapper

	lastActiveTime time.Time

	alive bool
}

func NewClient(role message.CRole, conn *websocket.Conn) *Client {
	out := make(chan message.Wrapper, 256) // buffer 256 send requests
	in := make(chan message.Wrapper, 256)  // buffer 256 send requests
	return &Client{
		conn:  conn,
		Out:   out,
		In:    in,
		role:  role,
		alive: true,
	}
}

func (cl *Client) Role() message.CRole {
	return cl.role
}

func (cl *Client) Alive() bool {
	return cl.alive
}

func (cl *Client) Start() {
	cl.conn.SetPongHandler(func(appData string) error {
		cl.lastActiveTime = time.Now()
		return nil
	})

	// periodically ping client
	go func() {
		for _ = range time.Tick(cfg.SERVER_PING_INTERVAL) {
			cl.conn.WriteControl(websocket.PingMessage, emptyByteArray, time.Time{})
			if time.Now().Sub(cl.lastActiveTime) > cfg.SERVER_DISCONNECTED_THRESHHOLD {
				cl.alive = false
				cl.conn.Close()
				log.Printf("Closing client role: %s due to inactive", cl.Role())
				return
			}
		}
	}()

	// Receive message coroutine
	go func() {
		for {
			msg, ok := <-cl.Out
			cl.lastActiveTime = time.Now()
			if ok {
				err := cl.conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Failed to boardcast to. Closing connection")
					cl.Close()
					return
				}
			} else {
				log.Printf("Failed to get message from channel")
				cl.Close()
				return
			}
		}
	}()

	// Send message coroutine
	for {
		msg := message.Wrapper{}
		err := cl.conn.ReadJSON(&msg)
		if err == nil {
			cl.In <- msg // Will be handled in Room
		} else {
			log.Printf("Failed to read message. Closing connection: %s", err)
			cl.Close()
			return
		}
	}
}

func (cl *Client) Close() {
	log.Printf("Closing client")
	cl.conn.WriteControl(websocket.CloseMessage, emptyByteArray, time.Time{})
	time.Sleep(1 * time.Second) // wait for client to receive close message
	cl.alive = false
	cl.conn.Close()
}
