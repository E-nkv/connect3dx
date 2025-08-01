package hub

import "encoding/json"

type MessageType int
type WsStatus int

const (
	WS_STATUS_OK WsStatus = iota
	WS_STATUS_BAD_REQUEST
	WS_STATUS_SERVER_ERROR
	WS_STATUS_UNJOINABLE
	WS_STATUS_ENEMY_JOINED
	WS_STATUS_ENEMY_SENT_MOVE
	WS_STATUS_GAMEOVER_WON
	WS_STATUS_GAMEOVER_LOST
	WS_STATUS_GAMEOVER_DRAW
)
const (
	MESSAGE_TYPE_REGISTER_MOVE_2D MessageType = iota
	MESSAGE_TYPE_JOIN_MATCH_2D
	MESSAGE_TYPE_CREATE_MATCH_2D
	MESSAGE_TYPE_ABANDON_MATCH_2D
	MESSAGE_TYPE_ASK_DRAW_2D
)

type WsRequest struct {
	Type MessageType     `json:"type"`
	Body json.RawMessage `json:"body"`
	ID   string          `json:"id"`
}

type WsResponse struct {
	ReqID  string   `json:"req_id"`
	Status WsStatus `json:"status"`
	Body   any      `json:"body"`
}
