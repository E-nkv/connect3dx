package hub

import (
	"bytes"
	"connectx/src/core"
	"connectx/src/models"
	"connectx/utils"
	"encoding/json"
	"fmt"

	"sync"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Hub struct {
	UserConns         map[string]*websocket.Conn
	UserConnsMutex    sync.Mutex
	UserModel         *models.User
	MatchController2D *core.MatchController2D
	MatchController3D *core.MatchController3D
}

func NewHub() *Hub {
	return &Hub{
		UserConns:         make(map[string]*websocket.Conn),
		MatchController2D: core.NewMatchController2D(),
		MatchController3D: core.NewMatchController3D(),
		UserModel:         &models.User{},
	}
}

func (h *Hub) ProcessMessage(userID string, conn *websocket.Conn, msg []byte, mt int) {
	switch mt {
	case websocket.BinaryMessage:
		var req WsRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			fmt.Println("err unmarshaling json: ", err)
			conn.WriteMessage(websocket.TextMessage, []byte("err unmarshaling json: "+err.Error()))
			return
		}
		switch req.Type {
		case MESSAGE_TYPE_CREATE_MATCH_2D:
			h.HandleCreateMatch2D(userID, conn, req)
		case MESSAGE_TYPE_JOIN_MATCH_2D:
			h.HandleJoinMatch2D(userID, conn, req)
		case MESSAGE_TYPE_REGISTER_MOVE_2D:
			h.HandleRegisterMove2D(userID, conn, req)
		}
	default:
		fmt.Println("expected binary, got msg type: ", mt)
		conn.WriteMessage(websocket.TextMessage, []byte("invalid msg type. expected Binary"))
	}
}

func (hub *Hub) ListenFromUser(userID string, conn *websocket.Conn) error {
	hub.UserConnsMutex.Lock()
	hub.UserConns[userID] = conn
	hub.UserConnsMutex.Unlock()

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			//probably disconnected
			fmt.Println("error reading message: ", err)
			hub.UserConnsMutex.Lock()
			delete(hub.UserConns, userID)
			hub.UserConnsMutex.Unlock()
			return err
		}

		message = bytes.ReplaceAll(message, newline, space)

		hub.ProcessMessage(userID, conn, message, mt)
	}

}

func writeMessage(conn *websocket.Conn, status WsStatus, id string, body any) {
	resp := WsResponse{
		ReqID:  id,
		Status: status,
		Body:   body,
	}
	bs, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("err marshaling response: ", err)
		conn.WriteMessage(websocket.TextMessage, []byte("SERVER ERROR"))
	}
	conn.WriteMessage(websocket.BinaryMessage, bs)
}

func writeError(conn *websocket.Conn, status WsStatus, id string, msg string) {
	writeMessage(conn, status, id, utils.Object{"error": msg})
}
