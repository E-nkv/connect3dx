package core

import (
	"connectx/src/types"
	"testing"
)

func TestMatchController3D_CreateMatch(t *testing.T) {
	c := NewMatchController3D()
	p1ID := "player1"
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}

	matchID, err := c.CreateMatch(p1ID, opts)
	if err != nil {
		t.Fatalf("CreateMatch failed: %v", err)
	}

	if matchID == "" {
		t.Fatal("CreateMatch returned an empty match ID")
	}

	c.MatchesMutex.Lock()
	defer c.MatchesMutex.Unlock()
	if _, ok := c.Matches[matchID]; !ok {
		t.Fatal("Match was not created in the controller")
	}
}

func TestMatchController3D_CreateMatch_InvalidOptions(t *testing.T) {
	c := NewMatchController3D()
	p1ID := "player1"
	opts := MatchOpts3D{R: 1, C: 4, H: 4, A: 4, Starts1: true} // Invalid R

	_, err := c.CreateMatch(p1ID, opts)
	if err == nil {
		t.Fatal("Expected an error for invalid match options, but got nil")
	}
}

func TestMatchController3D_JoinMatch(t *testing.T) {
	c := NewMatchController3D()
	p1ID := "player1"
	p2ID := "player2"
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}

	matchID, _ := c.CreateMatch(p1ID, opts)

	// Test joining a valid match
	match, isFirst, err := c.JoinMatch(p2ID, matchID)
	if err != nil {
		t.Fatalf("JoinMatch failed: %v", err)
	}
	if !isFirst {
		t.Fatal("Expected isFirst to be true for the first time joiner")
	}
	if match.P2.ID != p2ID {
		t.Fatalf("P2.ID was not set correctly: got %s, want %s", match.P2.ID, p2ID)
	}
	if !match.Started {
		t.Fatal("Match should be started after the second player joins")
	}

	// Test joining a match that is already full
	_, _, err = c.JoinMatch("player3", matchID)
	if err == nil {
		t.Fatal("Expected an error when joining a full match, but got nil")
	}

	// Test joining a non-existent match
	_, _, err = c.JoinMatch(p2ID, "non-existent-match")
	if err == nil {
		t.Fatal("Expected an error when joining a non-existent match, but got nil")
	}
}

func TestMatchController3D_RegisterMove(t *testing.T) {
	c := NewMatchController3D()
	p1ID := "player1"
	p2ID := "player2"
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}
	matchID, _ := c.CreateMatch(p1ID, opts)
	c.JoinMatch(p2ID, matchID)

	// Test a valid move
	pl := types.RegisterMove3DPL{MatchID: matchID, Row: 0, Col: 0}
	_, _, err := c.RegisterMove(p1ID, pl)
	if err != nil {
		t.Fatalf("RegisterMove failed for a valid move: %v", err)
	}

	// Test a move for the wrong player
	_, _, err = c.RegisterMove(p1ID, pl)
	if err == nil {
		t.Fatal("Expected an error when the wrong player tries to move, but got nil")
	}
}

func TestMatchController3D_RegisterMove_NotFound(t *testing.T) {
	c := NewMatchController3D()
	p1ID := "player1"

	pl := types.RegisterMove3DPL{MatchID: "non-existent-match", Row: 0, Col: 0}
	_, _, err := c.RegisterMove(p1ID, pl)
	if err == nil {
		t.Fatal("Expected an error when registering a move for a non-existent match, but got nil")
	}
}