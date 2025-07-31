package core

import (
	"testing"
)

// Helper function to create a match instance for testing, bypassing the buggy NewMatch2D.
func createTestMatch(opts MatchOpts) *Match2D {
	board := make([][]Slot, opts.H)
	for i := 0; i < opts.H; i++ {
		board[i] = make([]Slot, opts.W)
	}
	return &Match2D{
		Opts:  opts,
		P1:    Player{ID: "p1"},
		P2:    Player{ID: "p2"},
		Board: board,
		Moves: make([]Move, 0),
	}
}

// TestDirection_OtherSide tests the simple direction inversion.
func TestDirection_OtherSide(t *testing.T) {
	dir := Direction{Row: 1, Col: -1}
	expected := Direction{Row: -1, Col: 1}
	if other := dir.OtherSide(); other != expected {
		t.Errorf("Expected OtherSide of %v to be %v, but got %v", dir, expected, other)
	}
}

// TestGetRow correctly identifies the next available row.
func TestGetRow(t *testing.T) {
	match := createTestMatch(MatchOpts{W: 3, H: 3})
	match.Board[2][1] = SLOT_PLAYER1
	match.Board[1][1] = SLOT_PLAYER2

	if row := match.getRow(0); row != 2 {
		t.Errorf("Expected row 2 for empty column 0, got %d", row)
	}
	if row := match.getRow(1); row != 0 {
		t.Errorf("Expected row 0 for partially filled column 1, got %d", row)
	}

	match.Board[0][1] = SLOT_PLAYER1
	if row := match.getRow(1); row != -1 {
		t.Errorf("Expected row -1 for full column 1, got %d", row)
	}
}

func TestGetVictoryLine(t *testing.T) {
	opts := MatchOpts{W: 7, H: 6, A: 4}

	t.Run("Horizontal Win", func(t *testing.T) {
		match := createTestMatch(opts)
		match.Board[5][1] = SLOT_PLAYER1
		match.Board[5][2] = SLOT_PLAYER1
		match.Board[5][3] = SLOT_PLAYER1
		match.Board[5][4] = SLOT_PLAYER1
		line := match.getVictoryLine(5, 2, Direction{Row: 0, Col: 1})
		if len(line) < 4 {
			t.Fatalf("Failed to detect horizontal win. Line was: %v", line)
		}
	})

	t.Run("Vertical Win", func(t *testing.T) {
		match := createTestMatch(opts)
		match.Board[2][2] = SLOT_PLAYER2
		match.Board[3][2] = SLOT_PLAYER2
		match.Board[4][2] = SLOT_PLAYER2
		match.Board[5][2] = SLOT_PLAYER2
		line := match.getVictoryLine(3, 2, Direction{Row: 1, Col: 0})
		if len(line) < 4 {
			t.Fatalf("Failed to detect vertical win. Line was: %v", line)
		}
	})

	t.Run("Diagonal Win", func(t *testing.T) {
		match := createTestMatch(opts)
		match.Board[2][2] = SLOT_PLAYER1
		match.Board[3][3] = SLOT_PLAYER1
		match.Board[4][4] = SLOT_PLAYER1
		match.Board[5][5] = SLOT_PLAYER1
		line := match.getVictoryLine(3, 3, Direction{Row: 1, Col: 1})
		if line == nil || len(line) < 4 {
			t.Fatalf("Failed to detect diagonal win. Line was: %v", line)
		}
	})

	t.Run("Anti-Diagonal Win", func(t *testing.T) {
		match := createTestMatch(opts)
		match.Board[5][1] = SLOT_PLAYER2
		match.Board[4][2] = SLOT_PLAYER2
		match.Board[3][3] = SLOT_PLAYER2
		match.Board[2][4] = SLOT_PLAYER2
		line := match.getVictoryLine(4, 2, Direction{Row: 1, Col: -1})
		if line == nil || len(line) < 4 {
			t.Fatalf("Failed to detect anti-diagonal win. Line was: %v", line)
		}
	})

	t.Run("No Win - Interrupted Line", func(t *testing.T) {
		match := createTestMatch(opts)
		match.Board[5][1] = SLOT_PLAYER1
		match.Board[5][2] = SLOT_PLAYER1
		match.Board[5][3] = SLOT_PLAYER2 // Opponent's piece breaks the line
		match.Board[5][4] = SLOT_PLAYER1
		line := match.getVictoryLine(5, 2, Direction{Row: 0, Col: 1})
		if line != nil {
			t.Fatalf("Incorrectly found a winning line where it was interrupted: %v", line)
		}
	})

	t.Run("No Win - Not Enough Pieces", func(t *testing.T) {
		match := createTestMatch(opts)
		match.Board[5][1] = SLOT_PLAYER1
		match.Board[5][2] = SLOT_PLAYER1
		match.Board[5][3] = SLOT_PLAYER1
		line := match.getVictoryLine(5, 2, Direction{Row: 0, Col: 1})
		if line != nil {
			t.Fatalf("Incorrectly found a winning line with only 3 pieces: %v", line)
		}
	})

}

// TestIsGameover tests the top-level game over logic.
func TestIsGameover(t *testing.T) {
	opts := MatchOpts{W: 7, H: 6, A: 4}

	t.Run("Game Not Over", func(t *testing.T) {
		match := createTestMatch(opts)
		match.Board[5][0] = SLOT_PLAYER1
		match.Moves = []Move{{Col: 0}}
		if res := match.isGameover(5, 0); res != nil {
			t.Errorf("Expected no gameover, but got %v", res)
		}
	})
	t.Run("Win - Many Lines", func(t *testing.T) {
		match := createTestMatch(opts)
		//h win
		match.Board[5][1] = SLOT_PLAYER1
		match.Board[5][2] = SLOT_PLAYER1
		match.Board[5][3] = SLOT_PLAYER1
		match.Board[5][4] = SLOT_PLAYER1
		match.Board[5][5] = SLOT_PLAYER1
		match.Board[5][6] = SLOT_PLAYER1
		//v win
		match.Board[4][3] = SLOT_PLAYER1
		match.Board[3][3] = SLOT_PLAYER1
		match.Board[2][3] = SLOT_PLAYER1
		match.Board[1][3] = SLOT_PLAYER1
		res := match.isGameover(5, 3)
		if res["resType"] != RESULT_TYPE_WON {
			t.Fatal("expected to win")
		}

		t.Log("lines are: ", res["lines"])
	})
	t.Run("Game Won", func(t *testing.T) {
		t.Run("vertical", func(t *testing.T) {
			match := createTestMatch(opts)
			match.Board[4][3] = SLOT_PLAYER1
			match.Board[3][3] = SLOT_PLAYER1
			match.Board[2][3] = SLOT_PLAYER1
			match.Board[1][3] = SLOT_PLAYER1
			res := match.isGameover(4, 3)
			if res["resType"] != RESULT_TYPE_WON {
				t.Error("expected win")
			}
			t.Log("vertical lines are: ", res["lines"])
		})
		t.Run("horizontal", func(t *testing.T) {
			match := createTestMatch(opts)
			match.Board[0][1] = SLOT_PLAYER1
			match.Board[0][2] = SLOT_PLAYER1
			match.Board[0][3] = SLOT_PLAYER1
			match.Board[0][4] = SLOT_PLAYER1
			res := match.isGameover(0, 1)
			if res["resType"] != RESULT_TYPE_WON {
				t.Error("expected win")
			}
			t.Log("horizontal lines are: ", res["lines"])
		})
		t.Run("hv", func(t *testing.T) {
			match := createTestMatch(opts)

			match.Board[3][3] = SLOT_PLAYER1
			match.Board[2][3] = SLOT_PLAYER1
			match.Board[1][3] = SLOT_PLAYER1
			match.Board[0][3] = SLOT_PLAYER1

			match.Board[0][0] = SLOT_PLAYER1
			match.Board[0][1] = SLOT_PLAYER1
			match.Board[0][2] = SLOT_PLAYER1

			res := match.isGameover(0, 3)
			if res["resType"] != RESULT_TYPE_WON {
				t.Error("expected win")
			}
			t.Log("horizontal lines are: ", res["lines"])
		})

	})

	t.Run("Game is a Draw", func(t *testing.T) {
		drawOpts := MatchOpts{W: 2, H: 2, A: 3}
		match := createTestMatch(drawOpts)
		// Fill board in a draw pattern
		match.Board[0][0], match.Board[0][1] = SLOT_PLAYER1, SLOT_PLAYER2
		match.Board[1][0], match.Board[1][1] = SLOT_PLAYER2, SLOT_PLAYER1
		match.Moves = make([]Move, 4) // Fill moves history

		// The last move at (0,0) does not create a win, but fills the board.
		res := match.isGameover(0, 0)
		if res == nil {
			t.Fatal("Expected a draw result, but got nil")
		}
		if res["resType"] != RESULT_TYPE_DRAW {
			t.Errorf("Expected result type DRAW, but got %v", res["resType"])
		}
	})
}

// TestRegisterMove provides a comprehensive test for the main game logic function.
func TestRegisterMove(t *testing.T) {
	p1ID := "player1"
	p2ID := "player2"
	opts := MatchOpts{W: 7, H: 6, A: 4, Starts1: true}
	// Manually create match to avoid buggy constructor
	match := createTestMatch(opts)
	match.P1.ID = p1ID
	match.P2.ID = p2ID

	t.Run("Successful move", func(t *testing.T) {
		res, err := match.RegisterMove(Move{Col: 3}, p1ID)
		if err != nil {
			t.Fatalf("Expected valid move, got error: %v", err)
		}
		if res != nil {
			t.Fatalf("Expected no game over result on first move, got: %v", res)
		}
		if match.Board[5][3] != SLOT_PLAYER1 {
			t.Error("Board not updated correctly after P1 move")
		}
		if len(match.Moves) != 1 {
			t.Error("Move history not updated correctly")
		}
	})

	t.Run("Move by wrong player", func(t *testing.T) {
		_, err := match.RegisterMove(Move{Col: 0}, p1ID) // P1 tries to move again
		if err == nil || err.Error() != "not your turn" {
			t.Errorf("Expected 'not your turn' error, got: %v", err)
		}
	})

	t.Run("Move in invalid column (out of bounds)", func(t *testing.T) {
		// It's P2's turn.
		_, err := match.RegisterMove(Move{Col: -1}, p2ID)
		if err == nil || err.Error() != "invalid column" {
			t.Errorf("Expected 'invalid column' error for negative col, got: %v", err)
		}
		_, err = match.RegisterMove(Move{Col: 7}, p2ID) // W is 7, so max index is 6
		if err == nil || err.Error() != "invalid column" {
			t.Errorf("Expected 'invalid column' error for out of bounds col, got: %v", err)
		}
	})

	t.Run("Move resulting in a win", func(t *testing.T) {
		winMatch := createTestMatch(opts)
		winMatch.P1.ID, winMatch.P2.ID = p1ID, p2ID
		// Setup board for a win
		_, _ = winMatch.RegisterMove(Move{Col: 0}, p1ID)
		_, _ = winMatch.RegisterMove(Move{Col: 1}, p2ID)
		_, _ = winMatch.RegisterMove(Move{Col: 0}, p1ID)
		_, _ = winMatch.RegisterMove(Move{Col: 2}, p2ID)
		_, _ = winMatch.RegisterMove(Move{Col: 0}, p1ID)
		_, _ = winMatch.RegisterMove(Move{Col: 3}, p2ID)
		// Final winning move for P1 (vertical)
		res, err := winMatch.RegisterMove(Move{Col: 0}, p1ID)

		if err != nil {
			t.Fatalf("Winning move produced an error: %v", err)
		}
		if res == nil {
			t.Fatal("Winning move did not return a gameover result")
		}
		if res["resType"] != RESULT_TYPE_WON {
			t.Errorf("Expected result type WON, got %v", res["resType"])
		}
	})
}
