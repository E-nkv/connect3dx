package api

import (
	"bytes"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
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

func (h *Hub) ProcessMessage(userID string, conn *websocket.Conn, msg []byte, mt int) {
	switch mt {
	case websocket.BinaryMessage:
		fmt.Println("b msg received!", string(msg))
		conn.WriteMessage(websocket.TextMessage, []byte("b msg received!"))
	default:
		fmt.Println("unknown msg type: ", mt)
	}
}

func (hub *Hub) ListenFromUser(userID string, conn *websocket.Conn) error {
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexp close error: %v\n", err)
				return err
			}
			log.Printf("error reading message: %v\n", err)
			return err
		}

		message = bytes.ReplaceAll(message, newline, space)

		hub.ProcessMessage(userID, conn, message, mt)
	}

}
