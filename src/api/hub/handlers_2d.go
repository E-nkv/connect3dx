package hub

import (
	"connectx/src/core"
	"encoding/json"

	"github.com/gorilla/websocket"
)

func (h *Hub) HandleCreateMatch2D(userID string, conn *websocket.Conn, Req WsRequest) {
	var opts core.MatchOpts
	err := json.Unmarshal(Req.Body, &opts)
	if err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, Req.ID, "Invalid Request Body")
		return
	}
	id, err := h.MatchController2D.CreateMatch(userID, opts)
	if err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, Req.ID, err.Error())
		return
	}
	resp := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	writeMessage(conn, WS_STATUS_OK, Req.ID, resp)
}
