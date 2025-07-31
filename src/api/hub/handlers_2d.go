package hub

import (
	"connectx/src/core"
	"connectx/src/errs"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

func (h *Hub) HandleCreateMatch2D(userID string, conn *websocket.Conn, req WsRequest) {
	var opts core.MatchOpts
	err := json.Unmarshal(req.Body, &opts)
	if err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, "Invalid Request Body")
		return
	}
	id, err := h.MatchController2D.CreateMatch(userID, opts)
	if err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, err.Error())
		return
	}
	resp := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	writeMessage(conn, WS_STATUS_OK, req.ID, resp)
}

type JoinMatchPL struct {
	MatchID string `json:"match_id"`
}

func (h *Hub) HandleJoinMatch2D(userID string, conn *websocket.Conn, req WsRequest) {
	var pl JoinMatchPL
	if err := json.Unmarshal(req.Body, &pl); err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, "Invalid Request Body")
		return
	}
	match, isFirstTimeJoiner, err := h.MatchController2D.JoinMatch(userID, pl.MatchID)
	if err != nil {
		switch err {
		case errs.ErrNotFound:
			writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, "Match not found")
		case errs.ErrUnjoinable:
			writeError(conn, WS_STATUS_UNJOINABLE, req.ID, "Match unjoinable")
		}
		writeError(conn, WS_STATUS_SERVER_ERROR, req.ID, "Server error")
		return
	}

	var enemyConn *websocket.Conn
	var playerDTO *PlayerDTO
	if isFirstTimeJoiner {
		enemyID := match.GetEnemyID(userID)
		ec, ok := h.UserConns[enemyID]
		if !ok {
			//TODO: contiously check player1 connection to send him the enemy joined message
			return
		}
		enemyConn = ec
		playerData, err := h.UserModel.GetUserDTO(userID)
		if err != nil {
			//handle err
		}
		playerDTO = playerData

	}
	match.StartedAt = time.Now()
	go writeMessage(enemyConn, WS_STATUS_ENEMY_JOINED, req.ID, playerDTO)
	writeMessage(conn, WS_STATUS_OK, req.ID, match)
}
