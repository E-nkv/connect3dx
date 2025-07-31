package hub

import (
	"connectx/src/core"
	"encoding/json"
	"time"
)

type MessageType int
type WsStatus int

const (
	WS_STATUS_OK WsStatus = iota
	WS_STATUS_BAD_REQUEST
	WS_STATUS_SERVER_ERROR
	WS_STATUS_UNJOINABLE
	WS_STATUS_ENEMY_JOINED
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

type PlayerDTO struct {
	ID       string `json:"id"`
	TimeLeft int64
	Nick     string
	ImgURL   string
}
type Match2DDTO struct {
	Board     [][]core.Slot `json:"board"`
	P1        PlayerDTO
	P2        PlayerDTO
	Opts      core.MatchOpts
	Moves     []core.Move
	StartedAt time.Time
}
