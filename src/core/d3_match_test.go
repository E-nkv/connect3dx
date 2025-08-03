package core

import (
	"testing"
	"time"
)

func TestMatch3D_RegisterMove_Win(t *testing.T) {
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}
	match, _ := NewMatch3D("p1", "p2", opts)
	match.Started = true

	moves := []Move3D{
		{Row: 0, Col: 0}, {Row: 1, Col: 0},
		{Row: 0, Col: 0}, {Row: 1, Col: 0},
		{Row: 0, Col: 0}, {Row: 1, Col: 0},
		{Row: 0, Col: 0},
	}

	var res GameoverResult3D
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

func TestMatch3D_RegisterMove_NotYourTurn(t *testing.T) {
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}
	match, _ := NewMatch3D("p1", "p2", opts)
	match.Started = true

	_, err := match.RegisterMove(Move3D{Row: 0, Col: 0, RegisteredAt: time.Now()}, "p2")
	if err == nil {
		t.Fatal("expected an error for moving out of turn, but got nil")
	}
}

func TestMatch3D_RegisterMove_StickFull(t *testing.T) {
	opts := MatchOpts3D{R: 3, C: 3, H: 3, A: 3, Starts1: true}
	match, err := NewMatch3D("p1", "p2", opts)
	if err != nil {
		t.Fatal("unexpected err in match creation: ", err)
	}
	match.Started = true

	_, err = match.RegisterMove(Move3D{Row: 0, Col: 0}, "p1")
	if err != nil {
		t.Fatalf("unexpected error on move 1: %v", err)
	}
	_, err = match.RegisterMove(Move3D{Row: 0, Col: 0}, "p2")
	if err != nil {
		t.Fatalf("unexpected error on move 2: %v", err)
	}
	_, err = match.RegisterMove(Move3D{Row: 0, Col: 0}, "p1")
	if err != nil {
		t.Fatalf("unexpected error on move 3: %v", err)
	}
	_, err = match.RegisterMove(Move3D{Row: 0, Col: 0}, "p2")
	if err == nil {
		t.Fatal("expected an error for full stick, but got nil")
	}
}

func TestMatch3D_RegisterMove_InvalidCoordinates(t *testing.T) {
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}
	match, _ := NewMatch3D("p1", "p2", opts)
	match.Started = true

	_, err := match.RegisterMove(Move3D{Row: 4, Col: 0}, "p1")
	if err == nil {
		t.Fatal("expected an error for invalid row, but got nil")
	}

	_, err = match.RegisterMove(Move3D{Row: 0, Col: 4}, "p1")
	if err == nil {
		t.Fatal("expected an error for invalid col, but got nil")
	}
}

func TestMatch3D_RegisterMove_Gameover(t *testing.T) {
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}
	match, _ := NewMatch3D("p1", "p2", opts)
	match.Started = true
	match.Gameover = true

	_, err := match.RegisterMove(Move3D{Row: 0, Col: 0}, "p1")
	if err == nil {
		t.Fatal("expected an error for making a move on a game that is over, but got nil")
	}
}
