package hub_test

import (
	"connectx/src/api"
	"connectx/src/core"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Re-define constants and types for black-box testing
const (
	WS_STATUS_OK = iota
	WS_STATUS_BAD_REQUEST
	WS_STATUS_SERVER_ERROR
	WS_STATUS_UNJOINABLE
	WS_STATUS_ENEMY_JOINED
)
const (
	MESSAGE_TYPE_REGISTER_MOVE_2D = iota
	MESSAGE_TYPE_JOIN_MATCH_2D
	MESSAGE_TYPE_CREATE_MATCH_2D
)

type WsRequest struct {
	Type int             `json:"type"`
	Body json.RawMessage `json:"body"`
	ID   string          `json:"id"`
}

type WsResponse struct {
	ReqID  string          `json:"req_id"`
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

// Helper function to create a test server and a client connection
func setupTestServer(_ *testing.T) (*httptest.Server, *api.App) {
	app := api.NewApp()
	server := httptest.NewServer(http.HandlerFunc(app.HandleWs))
	return server, app
}

// Helper function to create a websocket client
func newWsClient(t *testing.T, serverURL string, userID string) *websocket.Conn {
	url := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws"
	header := http.Header{}
	header.Add("Cookie", "token="+userID)
	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("failed to connect to ws: %v", err)
	}
	return conn
}

func TestCreateAndJoinMatch(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Client 1 creates a match
	p1ID := "player1"
	p1Conn := newWsClient(t, server.URL, p1ID)
	defer p1Conn.Close()

	createReq := WsRequest{
		ID:   "1",
		Type: MESSAGE_TYPE_CREATE_MATCH_2D,
		Body: json.RawMessage(`{"w": 7, "h": 6, "a": 4, "starts1": true, "t0": 60, "td": 0}`),
	}
	createReqBytes, _ := json.Marshal(createReq)
	if err := p1Conn.WriteMessage(websocket.BinaryMessage, createReqBytes); err != nil {
		t.Fatalf("p1 failed to send create message: %v", err)
	}

	// Read response for create match
	var createResp WsResponse
	if err := p1Conn.ReadJSON(&createResp); err != nil {
		t.Fatalf("p1 failed to read create response: %v", err)
	}

	if createResp.Status != WS_STATUS_OK {
		t.Fatalf("expected status OK, got %d", createResp.Status)
	}

	var createBody struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createResp.Body, &createBody); err != nil {
		t.Fatalf("failed to unmarshal create response body: %v", err)
	}
	matchID := createBody.ID

	// Client 2 joins the match
	p2ID := "player2"
	p2Conn := newWsClient(t, server.URL, p2ID)
	defer p2Conn.Close()

	joinReq := WsRequest{
		ID:   "2",
		Type: MESSAGE_TYPE_JOIN_MATCH_2D,
		Body: json.RawMessage(`{"match_id": "` + matchID + `"}`),
	}
	joinReqBytes, _ := json.Marshal(joinReq)
	if err := p2Conn.WriteMessage(websocket.BinaryMessage, joinReqBytes); err != nil {
		t.Fatalf("p2 failed to send join message: %v", err)
	}

	// P2 reads their own successful join response
	var joinResp WsResponse
	if err := p2Conn.ReadJSON(&joinResp); err != nil {
		t.Fatalf("p2 failed to read join response: %v", err)
	}
	if joinResp.Status != WS_STATUS_OK {
		t.Fatalf("p2 expected status OK, got %d", joinResp.Status)
	}

	// P1 should receive an ENEMY_JOINED message
	var enemyJoinedResp WsResponse
	p1Conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	if err := p1Conn.ReadJSON(&enemyJoinedResp); err != nil {
		t.Fatalf("p1 failed to read enemy joined response: %v", err)
	}

	if enemyJoinedResp.Status != WS_STATUS_ENEMY_JOINED {
		t.Fatalf("p1 expected status ENEMY_JOINED, got %d", enemyJoinedResp.Status)
	}

	var enemyData core.PlayerDTO
	if err := json.Unmarshal(enemyJoinedResp.Body, &enemyData); err != nil {
		t.Fatalf("failed to unmarshal enemy data: %v", err)
	}

	if enemyData.ID != p2ID {
		t.Fatalf("expected enemy ID to be %s, got %s", p2ID, enemyData.ID)
	}
}

func TestRegisterMove(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// P1 creates a match
	p1ID := "player1"
	p1Conn := newWsClient(t, server.URL, p1ID)
	defer p1Conn.Close()

	createReq := WsRequest{
		ID:   "1",
		Type: MESSAGE_TYPE_CREATE_MATCH_2D,
		Body: json.RawMessage(`{"w": 7, "h": 6, "a": 4, "starts1": true, "t0": 60, "td": 0}`),
	}
	createReqBytes, _ := json.Marshal(createReq)
	p1Conn.WriteMessage(websocket.BinaryMessage, createReqBytes)
	var createResp WsResponse
	p1Conn.ReadJSON(&createResp)
	var createBody struct {
		ID string `json:"id"`
	}
	json.Unmarshal(createResp.Body, &createBody)
	matchID := createBody.ID

	// P2 joins the match
	p2ID := "player2"
	p2Conn := newWsClient(t, server.URL, p2ID)
	defer p2Conn.Close()

	joinReq := WsRequest{
		ID:   "2",
		Type: MESSAGE_TYPE_JOIN_MATCH_2D,
		Body: json.RawMessage(`{"match_id": "` + matchID + `"}`),
	}
	joinReqBytes, _ := json.Marshal(joinReq)
	p2Conn.WriteMessage(websocket.BinaryMessage, joinReqBytes)
	// Clear the two messages P2 and P1 get
	p2Conn.ReadJSON(&WsResponse{})
	p1Conn.ReadJSON(&WsResponse{})

	// P1 (who starts) sends a move
	moveReq := WsRequest{
		ID:   "3",
		Type: MESSAGE_TYPE_REGISTER_MOVE_2D,
		Body: json.RawMessage(`{"match_id": "` + matchID + `", "col": 3}`),
	}
	moveReqBytes, _ := json.Marshal(moveReq)
	if err := p1Conn.WriteMessage(websocket.BinaryMessage, moveReqBytes); err != nil {
		t.Fatalf("p1 failed to send move: %v", err)
	}

	// P1 gets an OK response
	var moveResp WsResponse
	if err := p1Conn.ReadJSON(&moveResp); err != nil {
		t.Fatalf("p1 failed to read move response: %v", err)
	}
	if moveResp.Status != WS_STATUS_OK {
		t.Fatalf("p1 expected status OK for move, got %d", moveResp.Status)
	}

	// P2 should receive an ENEMY_SENT_MOVE message
	const WS_STATUS_ENEMY_SENT_MOVE = 5
	var enemyMoveResp WsResponse
	p2Conn.SetReadDeadline(time.Now().Add(time.Second * 1))
	if err := p2Conn.ReadJSON(&enemyMoveResp); err != nil {
		t.Fatalf("p2 failed to read enemy move notification: %v", err)
	}

	if enemyMoveResp.Status != WS_STATUS_ENEMY_SENT_MOVE {
		t.Fatalf("p2 expected status ENEMY_SENT_MOVE, got %d", enemyMoveResp.Status)
	}

	var enemyMoveBody struct {
		Col int `json:"col"`
	}
	if err := json.Unmarshal(enemyMoveResp.Body, &enemyMoveBody); err != nil {
		t.Fatalf("p2 failed to unmarshal enemy move body: %v", err)
	}

	if enemyMoveBody.Col != 3 {
		t.Fatalf("p2 expected enemy move in col 3, got %d", enemyMoveBody.Col)
	}
}
