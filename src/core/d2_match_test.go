package core

import (
	"testing"
)

func TestMatch2D_RegisterMove_Win(t *testing.T) {
	opts := MatchOpts{W: 7, H: 6, A: 4, Starts1: true}
	match, _ := NewMatch2D("p1", "p2", opts)
	match.Started = true

	moves := []Move{
		{Col: 0}, {Col: 1},
		{Col: 0}, {Col: 1},
		{Col: 0}, {Col: 1},
		{Col: 0},
	}

	var res GameoverResult
	var err error
	for i, move := range moves {
		pid := "p1"
		if i%2 != 0 {
			pid = "p2"
		}
		res, err = match.RegisterMove(move, pid)
		if err != nil {
			t.Fatalf("unexpected error on move %d: %v", i, err)
		}
	}

	if res == nil {
		t.Fatal("expected game to be over, but it was not")
	}
	if res["resType"] != RESULT_TYPE_WON {
		t.Fatalf("expected result type to be WON, but got %v", res["resType"])
	}
}

func TestMatch2D_RegisterMove_InvalidColumn(t *testing.T) {
	opts := MatchOpts{W: 7, H: 6, A: 4, Starts1: true}
	match, _ := NewMatch2D("p1", "p2", opts)
	match.Started = true

	_, err := match.RegisterMove(Move{Col: 7}, "p1")
	if err == nil {
		t.Fatal("expected an error for invalid column, but got nil")
	}
}

func TestMatch2D_RegisterMove_ColumnFull(t *testing.T) {
	opts := MatchOpts{W: 3, H: 2, A: 3, Starts1: true}
	match, _ := NewMatch2D("p1", "p2", opts)
	match.Started = true

	_, err := match.RegisterMove(Move{Col: 0}, "p1")
	if err != nil {
		t.Fatalf("unexpected error on move 1: %v", err)
	}
	_, err = match.RegisterMove(Move{Col: 0}, "p2")
	if err != nil {
		t.Fatalf("unexpected error on move 2: %v", err)
	}
	_, err = match.RegisterMove(Move{Col: 0}, "p1")
	if err == nil {
		t.Fatal("expected an error for full column, but got nil")
	}
}

func TestMatch2D_RegisterMove_NotYourTurn(t *testing.T) {
	opts := MatchOpts{W: 7, H: 6, A: 4, Starts1: true}
	match, _ := NewMatch2D("p1", "p2", opts)
	match.Started = true

	_, err := match.RegisterMove(Move{Col: 0}, "p2")
	if err == nil {
		t.Fatal("expected an error for moving out of turn, but got nil")
	}
}

func TestMatch2D_RegisterMove_Gameover(t *testing.T) {
	opts := MatchOpts{W: 7, H: 6, A: 4, Starts1: true}
	match, _ := NewMatch2D("p1", "p2", opts)
	match.Started = true
	match.Gameover = true

	_, err := match.RegisterMove(Move{Col: 0}, "p1")
	if err == nil {
		t.Fatal("expected an error for making a move on a game that is over, but got nil")
	}
}
