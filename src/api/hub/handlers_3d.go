package hub

import (
	"connectx/src/core"
	"connectx/src/errs"
	"connectx/src/types"
	"connectx/utils"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func (h *Hub) HandleCreateMatch3D(userID string, conn *websocket.Conn, req WsRequest) {
	var opts core.MatchOpts3D
	err := json.Unmarshal(req.Body, &opts)
	if err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, "Invalid Request Body")
		return
	}
	id, err := h.MatchController3D.CreateMatch(userID, opts)
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

func (h *Hub) HandleJoinMatch3D(userID string, conn *websocket.Conn, req WsRequest) {
	var pl types.JoinMatchPL
	if err := json.Unmarshal(req.Body, &pl); err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, "Invalid Request Body")
		return
	}
	match, isFirstTimeJoiner, err := h.MatchController3D.JoinMatch(userID, pl.MatchID)
	if err != nil {
		switch err {
		case errs.ErrNotFound:
			writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, "Match not found")
		case errs.ErrUnjoinable:
			writeError(conn, WS_STATUS_UNJOINABLE, req.ID, "Match unjoinable")
		default:
			writeError(conn, WS_STATUS_SERVER_ERROR, req.ID, "Server error")
		}
		return
	}

	if isFirstTimeJoiner {
		enemyID := match.GetEnemyID(userID)

		h.UserConnsMutex.Lock()
		enemyConn, ok := h.UserConns[enemyID]
		h.UserConnsMutex.Unlock()

		if !ok {
			// Player 1 is not connected. We can't notify them.
			// The join for player 2 will succeed, and when player 1 reconnects,
			// they should fetch the latest match state.
			// For now, we just can't send the ENEMY_JOINED message.
		} else {
			playerData, err := h.UserModel.GetUserDTO(userID)
			if err != nil {
				writeError(conn, WS_STATUS_SERVER_ERROR, req.ID, "Could not retrieve joining player's data")
				// Note: The player has technically joined the match state in the controller.
				// A robust implementation would revert this. For now, we abort the handler.
				return
			}
			go writeMessage(enemyConn, WS_STATUS_ENEMY_JOINED, req.ID, playerData)
		}
	}

	writeMessage(conn, WS_STATUS_OK, req.ID, match)
}

func (h *Hub) HandleRegisterMove3D(userID string, conn *websocket.Conn, req WsRequest) {
	var body types.RegisterMovePL
	if err := json.Unmarshal(req.Body, &body); err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, "invalid move payload")
		return
	}
	m, res, err := h.MatchController3D.RegisterMove(userID, body)
	if err != nil {
		writeError(conn, WS_STATUS_BAD_REQUEST, req.ID, err.Error())
		return
	}
	enemyID := m.GetEnemyID(userID)
	h.UserConnsMutex.Lock()
	enemyConn, isEnemyConnected := h.UserConns[enemyID]
	h.UserConnsMutex.Unlock()

	switch {
	case res == nil:
		//normal move
		go writeMessage(conn, WS_STATUS_OK, req.ID, nil)
		if isEnemyConnected {
			writeMessage(enemyConn, WS_STATUS_ENEMY_SENT_MOVE, "-1", utils.Object{
				"col":          body.Col,
				"time_left_p1": m.P1.TimeLeft,
				"time_left_p2": m.P2.TimeLeft,
			})
		}
	case res["resType"] == core.RESULT_TYPE_WON:
		//winning move

		b := utils.Object{
			"col":          body.Col,
			"time_left_p1": m.P1.TimeLeft,
			"time_left_p2": m.P2.TimeLeft,
			"lines":        res["lines"],
		}
		go writeMessage(conn, WS_STATUS_GAMEOVER_WON, req.ID, b)
		if isEnemyConnected {
			writeMessage(enemyConn, WS_STATUS_GAMEOVER_LOST, "-1", b)
		}
	case res["resType"] == core.RESULT_TYPE_DRAW:
		//drawing move

		b := utils.Object{
			"col":          body.Col,
			"time_left_p1": m.P1.TimeLeft,
			"time_left_p2": m.P2.TimeLeft,
		}
		go writeMessage(conn, WS_STATUS_GAMEOVER_DRAW, req.ID, b)
		if isEnemyConnected {
			writeMessage(enemyConn, WS_STATUS_GAMEOVER_DRAW, "-1", b)
		}
	default:
		fmt.Printf("unexpected scenario in handleRegisterMove.. \n\tres is: %+v\n\tand match is: %+v\n", res, m)
		writeError(conn, WS_STATUS_SERVER_ERROR, req.ID, "unexpected scenario in HandleRegisterMove")
	}

}
