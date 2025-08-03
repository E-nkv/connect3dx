package core

import (
	"testing"
)

// TestDirection_OtherSide tests the simple direction inversion.
func TestDirection_OtherSide(t *testing.T) {
	dir := Direction{Row: 1, Col: -1}
	expected := Direction{Row: -1, Col: 1}
	if other := dir.OtherSide(); other != expected {
		t.Errorf("Expected OtherSide of %v to be %v, but got %v", dir, expected, other)
	}
}

func TestGetRow(t *testing.T) {
	match, err := NewMatch2D("p1", "p2", MatchOpts{W: 3, H: 3, A: 3})
	if err != nil {
		t.Fatal("Failed to create match:", err)
	}
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
		match, _ := NewMatch2D("p1", "p2", opts)
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
		match, _ := NewMatch2D("p1", "p2", opts)
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
		match, _ := NewMatch2D("p1", "p2", opts)
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
		match, _ := NewMatch2D("p1", "p2", opts)
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
		match, _ := NewMatch2D("p1", "p2", opts)
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
		match, _ := NewMatch2D("p1", "p2", opts)
		match.Board[5][1] = SLOT_PLAYER1
		match.Board[5][2] = SLOT_PLAYER1
		match.Board[5][3] = SLOT_PLAYER1
		line := match.getVictoryLine(5, 2, Direction{Row: 0, Col: 1})
		if line != nil {
			t.Fatalf("Incorrectly found a winning line with only 3 pieces: %v", line)
		}
	})

}

func TestIsGameover(t *testing.T) {
	opts := MatchOpts{W: 7, H: 6, A: 4}

	t.Run("Game Not Over", func(t *testing.T) {
		match, _ := NewMatch2D("p1", "p2", opts)
		match.Board[5][0] = SLOT_PLAYER1
		match.Moves = []Move{{Col: 0}}
		if res := match.isGameover(5, 0); res != nil {
			t.Errorf("Expected no gameover, but got %v", res)
		}
	})
	t.Run("Win - Many Lines", func(t *testing.T) {
		match, _ := NewMatch2D("p1", "p2", opts)
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
			match, _ := NewMatch2D("p1", "p2", opts)
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
			match, _ := NewMatch2D("p1", "p2", opts)
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
			match, _ := NewMatch2D("p1", "p2", opts)

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
		// A 3x3 board with A=3 can result in a draw.
		opts := MatchOpts{W: 3, H: 3, A: 3, Starts1: true}
		match, err := NewMatch2D("p1", "p2", opts)
		if err != nil {
			t.Fatalf("Failed to create match for draw test: %v", err)
		}

		// A sequence of moves that results in a draw
		// Board state:
		// p1 p2 p1
		// p1 p2 p1
		// p2 p1 p2
		moves := []int{0, 1, 0, 1, 2, 2, 1, 0, 2}
		players := []string{"p1", "p2", "p1", "p2", "p1", "p2", "p1", "p2", "p1"}

		var res GameoverResult
		for i, col := range moves {
			// Figure out correct player based on turn
			player := "p1"
			if i%2 != 0 {
				player = "p2"
			}
			if !opts.Starts1 {
				if player == "p1" {
					player = "p2"
				} else {
					player = "p1"
				}
			}

			res, err = match.RegisterMove(Move{Col: col}, player)
			if err != nil {
				t.Fatalf("Move %d by %s in col %d failed: %v", i+1, players[i], col, err)
			}
		}

		if res == nil {
			t.Fatal("Expected a draw result on the final move, but got nil")
		}
		if res["resType"] != RESULT_TYPE_DRAW {
			t.Errorf("Expected result type DRAW, but got %v", res["resType"])
		}
	})
}

func TestRegisterMove(t *testing.T) {
	p1ID := "player1"
	p2ID := "player2"
	opts := MatchOpts{W: 7, H: 6, A: 4, Starts1: true}
	match, err := NewMatch2D(p1ID, p2ID, opts)
	if err != nil {
		t.Fatal("Failed to create match:", err)
	}

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

	t.Run("Move in full column", func(t *testing.T) {
		match, err := NewMatch2D(p1ID, p2ID, MatchOpts{W: 2, H: 1, A: 2, Starts1: true})
		if err != nil {
			t.Fatalf("NewMatch2D failed unexpectedly: %v", err)
		}
		_, _ = match.RegisterMove(Move{Col: 0}, p1ID)
		_, _ = match.RegisterMove(Move{Col: 1}, p2ID)
		_, err = match.RegisterMove(Move{Col: 0}, p1ID)
		if err == nil || err.Error() != "invalid move. column is full" {
			t.Errorf("Expected 'invalid move. column is full' error, got: %v", err)
		}
	})

	t.Run("Move resulting in a win", func(t *testing.T) {
		winMatch, _ := NewMatch2D(p1ID, p2ID, opts)
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

	t.Run("Win on last move", func(t *testing.T) {
		opts := MatchOpts{W: 2, H: 2, A: 2, Starts1: true}
		match, _ := NewMatch2D(p1ID, p2ID, opts)
		// Moves:
		// p1 -> (1,0)
		// p2 -> (0,0)
		// p1 -> (1,1)
		// p2 -> (0,1) makes a horizontal line for p2
		_, _ = match.RegisterMove(Move{Col: 0}, p1ID)
		_, _ = match.RegisterMove(Move{Col: 0}, p2ID)
		_, _ = match.RegisterMove(Move{Col: 1}, p1ID)
		res, err := match.RegisterMove(Move{Col: 1}, p2ID)

		// Board state after moves:
		// row 0: p2, p2
		// row 1: p1, p1
		// p2's move at (0,1) creates a horizontal win for p2.
		// The board is also full.

		if err != nil {
			t.Fatalf("Winning move on full board produced an error: %v", err)
		}
		if res == nil {
			t.Fatal("Expected a win result, but got nil")
		}
		if res["resType"] != RESULT_TYPE_WON {
			t.Errorf("Expected result type WON, but got %v", res["resType"])
		}
	})
}

func TestNewMatch2D(t *testing.T) {
	p1ID := "player1"
	p2ID := "player2"

	t.Run("Valid Match Creation", func(t *testing.T) {
		opts := MatchOpts{W: 7, H: 6, A: 4, T0: 60000, TD: 10000}
		match, err := NewMatch2D(p1ID, p2ID, opts)
		if err != nil {
			t.Fatalf("NewMatch2D failed with valid options: %v", err)
		}
		if match == nil {
			t.Fatal("NewMatch2D returned nil with valid options")
		}
		if match.P1.ID != p1ID || match.P2.ID != p2ID {
			t.Errorf("Players not initialized correctly")
		}
		if match.Opts.W != 7 || match.Opts.H != 6 || match.Opts.A != 4 {
			t.Errorf("Match options not set correctly")
		}
		if len(match.Board) != 6 || len(match.Board[0]) != 7 {
			t.Errorf("Board dimensions are incorrect")
		}
		if match.P1.TimeLeft != 60000 || match.P2.TimeLeft != 60000 {
			t.Errorf("Player time not initialized correctly")
		}
	})

	t.Run("Invalid Match Options", func(t *testing.T) {
		// Zero width
		_, err := NewMatch2D(p1ID, p2ID, MatchOpts{W: 0, H: 6, A: 4})
		if err == nil {
			t.Error("Expected error for zero width, got nil")
		}
		// Negative height
		_, err = NewMatch2D(p1ID, p2ID, MatchOpts{W: 7, H: -1, A: 4})
		if err == nil {
			t.Error("Expected error for negative height, got nil")
		}
		// Alignment greater than dimensions
		_, err = NewMatch2D(p1ID, p2ID, MatchOpts{W: 5, H: 5, A: 6})
		if err == nil {
			t.Error("Expected error for alignment > width and height, got nil")
		}
	})
}

func TestGetCurrPlayer(t *testing.T) {
	p1ID := "p1"
	p2ID := "p2"
	opts := MatchOpts{W: 7, H: 6, A: 4, Starts1: true}
	match, _ := NewMatch2D(p1ID, p2ID, opts)

	// P1 starts
	if match.getCurrPlayer() != p1ID {
		t.Errorf("Expected P1 to be current player at start, got %s", match.getCurrPlayer())
	}

	// After 1 move, should be P2
	match.Moves = append(match.Moves, Move{})
	if match.getCurrPlayer() != p2ID {
		t.Errorf("Expected P2 to be current player after 1 move, got %s", match.getCurrPlayer())
	}

	// P2 starts
	match.Opts.Starts1 = false
	match.Moves = []Move{} // Reset moves
	if match.getCurrPlayer() != p2ID {
		t.Errorf("Expected P2 to be current player at start, got %s", match.getCurrPlayer())
	}

	// After 1 move, should be P1
	match.Moves = append(match.Moves, Move{})
	if match.getCurrPlayer() != p1ID {
		t.Errorf("Expected P1 to be current player after 1 move, got %s", match.getCurrPlayer())
	}
}

func TestGetEnemyID(t *testing.T) {
	match, _ := NewMatch2D("player-one", "player-two", MatchOpts{W: 7, H: 6, A: 4})

	if enemy := match.GetEnemyID("player-one"); enemy != "player-two" {
		t.Errorf("Expected enemy of player-one to be player-two, got %s", enemy)
	}
	if enemy := match.GetEnemyID("player-two"); enemy != "player-one" {
		t.Errorf("Expected enemy of player-two to be player-one, got %s", enemy)
	}
	if enemy := match.GetEnemyID("non-existent-player"); enemy != "" {
		t.Errorf("Expected empty string for non-existent player, got %s", enemy)
	}
}
