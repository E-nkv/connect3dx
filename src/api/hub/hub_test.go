package hub

import (
	"connectx/src/core"
	"connectx/src/types"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockDTOGetter is a mock implementation of the DTOGetter interface.
type mockDTOGetter struct{}

func (m *mockDTOGetter) GetUserDTO(userID string) (*core.PlayerDTO, error) {
	return &core.PlayerDTO{
		ID:   userID,
		Nick: "player",
	}, nil
}

// newTestHub creates a new Hub for testing purposes.
func newTestHub() *Hub {
	return NewHub(&mockDTOGetter{})
}

// newTestConn creates a new websocket connection for testing.
func newTestConn(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	// Use a channel to pass the server connection from the handler
	serverConnChan := make(chan *websocket.Conn)

	// Create a test server
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("failed to upgrade connection: %v", err)
			return
		}
		serverConnChan <- c
	}))

	// Convert http:// to ws://
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	clientConn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		s.Close()
		t.Fatalf("failed to connect to websocket: %v", err)
	}

	// Wait for the server to send us the connection
	serverConn := <-serverConnChan

	// The test server and connections will be closed by t.Cleanup()
	t.Cleanup(func() {
		clientConn.Close()
		serverConn.Close()
		s.Close()
	})

	return serverConn, clientConn
}

func TestHub_HandleCreateMatch2D(t *testing.T) {
	hub := newTestHub()
	serverConn, clientConn := newTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	p1ID := "player1"
	hub.UserConns[p1ID] = serverConn

	opts := core.MatchOpts{W: 7, H: 6, A: 4}
	body, _ := json.Marshal(opts)
	req := WsRequest{
		Type: MESSAGE_TYPE_CREATE_MATCH_2D,
		ID:   "1",
		Body: body,
	}
	reqBytes, _ := json.Marshal(req)

	hub.ProcessMessage(p1ID, serverConn, reqBytes, websocket.BinaryMessage)

	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var resp WsResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != WS_STATUS_OK {
		t.Errorf("expected status OK, got %v", resp.Status)
	}
	if resp.ReqID != "1" {
		t.Errorf("expected req_id 1, got %s", resp.ReqID)
	}
	var respBody struct {
		ID string `json:"id"`
	}
	bodyBytes, err := json.Marshal(resp.Body)
	if err != nil {
		t.Fatalf("failed to marshal response body: %v", err)
	}
	if err := json.Unmarshal(bodyBytes, &respBody); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if respBody.ID == "" {
		t.Error("expected a match ID, but it was empty")
	}
}

func TestHub_HandleJoinMatch2D(t *testing.T) {
	hub := newTestHub()
	p1Conn, p1ClientConn := newTestConn(t)
	p2Conn, p2ClientConn := newTestConn(t)
	defer p1Conn.Close()
	defer p1ClientConn.Close()
	defer p2Conn.Close()
	defer p2ClientConn.Close()

	p1ID := "player1"
	p2ID := "player2"
	hub.UserConns[p1ID] = p1Conn
	hub.UserConns[p2ID] = p2Conn

	opts := core.MatchOpts{W: 7, H: 6, A: 4}
	matchID, _ := hub.MatchController2D.CreateMatch(p1ID, opts)

	joinReq := types.JoinMatchPL{MatchID: matchID}
	body, _ := json.Marshal(joinReq)
	req := WsRequest{
		Type: MESSAGE_TYPE_JOIN_MATCH_2D,
		ID:   "2",
		Body: body,
	}
	reqBytes, _ := json.Marshal(req)

	hub.ProcessMessage(p2ID, p2Conn, reqBytes, websocket.BinaryMessage)

	// Check response to player 2
	_, msg, err := p2ClientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from p2: %v", err)
	}
	var resp WsResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != WS_STATUS_OK {
		t.Errorf("expected status OK for p2, got %v", resp.Status)
	}

	// Check message to player 1
	_, msg, err = p1ClientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from p1: %v", err)
	}
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != WS_STATUS_ENEMY_JOINED {
		t.Errorf("expected status ENEMY_JOINED for p1, got %v", resp.Status)
	}
}

func TestHub_HandleRegisterMove2D(t *testing.T) {
	hub := newTestHub()
	p1Conn, p1ClientConn := newTestConn(t)
	p2Conn, p2ClientConn := newTestConn(t)
	defer p1Conn.Close()
	defer p1ClientConn.Close()
	defer p2Conn.Close()
	defer p2ClientConn.Close()

	p1ID := "player1"
	p2ID := "player2"
	hub.UserConns[p1ID] = p1Conn
	hub.UserConns[p2ID] = p2Conn

	opts := core.MatchOpts{W: 7, H: 6, A: 4, Starts1: true}
	matchID, _ := hub.MatchController2D.CreateMatch(p1ID, opts)
	hub.MatchController2D.JoinMatch(p2ID, matchID)

	moveReq := types.RegisterMovePL{MatchID: matchID, Col: 0}
	body, _ := json.Marshal(moveReq)
	req := WsRequest{
		Type: MESSAGE_TYPE_REGISTER_MOVE_2D,
		ID:   "3",
		Body: body,
	}
	reqBytes, _ := json.Marshal(req)

	hub.ProcessMessage(p1ID, p1Conn, reqBytes, websocket.BinaryMessage)

	// Check response to player 1
	_, msg, err := p1ClientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from p1: %v", err)
	}
	var resp WsResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != WS_STATUS_OK {
		t.Errorf("expected status OK for p1, got %v", resp.Status)
	}

	// Check message to player 2
	_, msg, err = p2ClientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from p2: %v", err)
	}
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != WS_STATUS_ENEMY_SENT_MOVE {
		t.Errorf("expected status ENEMY_SENT_MOVE for p2, got %v", resp.Status)
	}
}

func TestHub_HandleCreateMatch3D(t *testing.T) {
	hub := newTestHub()
	serverConn, clientConn := newTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	p1ID := "player1"
	hub.UserConns[p1ID] = serverConn

	opts := core.MatchOpts3D{R: 4, C: 4, H: 4, A: 4}
	body, _ := json.Marshal(opts)
	req := WsRequest{
		Type: MESSAGE_TYPE_CREATE_MATCH_3D,
		ID:   "4",
		Body: body,
	}
	reqBytes, _ := json.Marshal(req)

	hub.ProcessMessage(p1ID, serverConn, reqBytes, websocket.BinaryMessage)

	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var resp WsResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != WS_STATUS_OK {
		t.Errorf("expected status OK, got %v", resp.Status)
	}
	if resp.ReqID != "4" {
		t.Errorf("expected req_id 4, got %s", resp.ReqID)
	}
	var respBody struct {
		ID string `json:"id"`
	}
	bodyBytes, err := json.Marshal(resp.Body)
	if err != nil {
		t.Fatalf("failed to marshal response body: %v", err)
	}
	if err := json.Unmarshal(bodyBytes, &respBody); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if respBody.ID == "" {
		t.Error("expected a match ID, but it was empty")
	}
}

func TestHub_HandleJoinMatch3D(t *testing.T) {
	hub := newTestHub()
	p1Conn, p1ClientConn := newTestConn(t)
	p2Conn, p2ClientConn := newTestConn(t)
	defer p1Conn.Close()
	defer p1ClientConn.Close()
	defer p2Conn.Close()
	defer p2ClientConn.Close()

	p1ID := "player1"
	p2ID := "player2"
	hub.UserConns[p1ID] = p1Conn
	hub.UserConns[p2ID] = p2Conn

	opts := core.MatchOpts3D{R: 4, C: 4, H: 4, A: 4}
	matchID, _ := hub.MatchController3D.CreateMatch(p1ID, opts)

	joinReq := types.JoinMatchPL{MatchID: matchID}
	body, _ := json.Marshal(joinReq)
	req := WsRequest{
		Type: MESSAGE_TYPE_JOIN_MATCH_3D,
		ID:   "5",
		Body: body,
	}
	reqBytes, _ := json.Marshal(req)

	hub.ProcessMessage(p2ID, p2Conn, reqBytes, websocket.BinaryMessage)

	// Check response to player 2
	_, msg, err := p2ClientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from p2: %v", err)
	}
	var resp WsResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != WS_STATUS_OK {
		t.Errorf("expected status OK for p2, got %v", resp.Status)
	}

	// Check message to player 1
	_, msg, err = p1ClientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from p1: %v", err)
	}
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != WS_STATUS_ENEMY_JOINED {
		t.Errorf("expected status ENEMY_JOINED for p1, got %v", resp.Status)
	}
}

func TestHub_HandleRegisterMove3D(t *testing.T) {
	hub := newTestHub()
	p1Conn, p1ClientConn := newTestConn(t)
	p2Conn, p2ClientConn := newTestConn(t)
	defer p1Conn.Close()
	defer p1ClientConn.Close()
	defer p2Conn.Close()
	defer p2ClientConn.Close()

	p1ID := "player1"
	p2ID := "player2"
	hub.UserConns[p1ID] = p1Conn
	hub.UserConns[p2ID] = p2Conn

	opts := core.MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}
	matchID, _ := hub.MatchController3D.CreateMatch(p1ID, opts)
	hub.MatchController3D.JoinMatch(p2ID, matchID)

	moveReq := types.RegisterMove3DPL{MatchID: matchID, Row: 0, Col: 0}
	body, _ := json.Marshal(moveReq)
	req := WsRequest{
		Type: MESSAGE_TYPE_REGISTER_MOVE_3D,
		ID:   "6",
		Body: body,
	}
	reqBytes, _ := json.Marshal(req)

	hub.ProcessMessage(p1ID, p1Conn, reqBytes, websocket.BinaryMessage)

	// Check response to player 1
	_, msg, err := p1ClientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from p1: %v", err)
	}
	var resp WsResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != WS_STATUS_OK {
		t.Errorf("expected status OK for p1, got %v", resp.Status)
	}

	// Check message to player 2
	_, msg, err = p2ClientConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message from p2: %v", err)
	}
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != WS_STATUS_ENEMY_SENT_MOVE {
		t.Errorf("expected status ENEMY_SENT_MOVE for p2, got %v", resp.Status)
	}
}


func TestHub_ListenFromUser_Disconnect(t *testing.T) {
	hub := newTestHub()
	serverConn, clientConn := newTestConn(t)

	p1ID := "player1"

	go func() {
		// This will block until the connection is closed
		hub.ListenFromUser(p1ID, serverConn)
	}()

	// Wait a moment to ensure the user is registered
	time.Sleep(100 * time.Millisecond)

	// Close the client connection to simulate a disconnect
	clientConn.Close()

	// Wait a moment to ensure the disconnect is processed
	time.Sleep(100 * time.Millisecond)

	hub.UserConnsMutex.Lock()
	defer hub.UserConnsMutex.Unlock()
	if _, ok := hub.UserConns[p1ID]; ok {
		t.Error("user connection was not removed after disconnect")
	}
}