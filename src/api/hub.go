package api

import (
	"bytes"
	"connectx/utils"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

type MessageType int
const (
	MESSAGE_TYPE_REGISTER_MOVE MessageType = iota
	MESSAGE_TYPE_JOIN_MATCH
	MESSAGE_TYPE_CREATE_MATCH
	MESSAGE_TYPE_ABANDON_MATCH
	MESSAGE_TYPE_ASK_DRAW

)
type Hub struct {
	Clients map[string]*websocket.Conn
}

func newHub() *Hub {
	return &Hub{
		Clients: make(map[string]*websocket.Conn),
	}
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Message struct {
	Type MessageType          `json:"type"`
	Body utils.Object `json:"body"`
}

func (h *Hub) ProcessMessage(userID string, conn *websocket.Conn, msg []byte, mt int) {
	switch mt {
	case websocket.BinaryMessage:
		var M Message
		if err := json.Unmarshal(msg, &M); err != nil {
			fmt.Println("err unmarshaling json: ", err)
			conn.WriteMessage(websocket.TextMessage, []byte("err unmarshaling json: "+err.Error()))
			return
		}
		switch M.Type {
			case 
		}
	default:
		fmt.Println("unknown msg type: ", mt)
		conn.WriteMessage(websocket.TextMessage, []byte("invalid msg type. expected Binary"))
	}
}

func (hub *Hub) ListenFromUser(userID string, conn *websocket.Conn) error {
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			//probably disconnected
			fmt.Println("error reading message: ", err)
			delete(hub.Clients, userID)
			return err
		}

		message = bytes.ReplaceAll(message, newline, space)

		hub.ProcessMessage(userID, conn, message, mt)
	}

}
