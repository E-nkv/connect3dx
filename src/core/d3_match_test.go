package core

import (
	"fmt"
	"testing"
)

// createTestMatch3D is a helper function to create a 3D match instance for testing.
func createTestMatch3D(opts MatchOpts3D) *Match3D {
	match, err := NewMatch3D("p1", "p2", opts)
	if err != nil {
		panic(fmt.Sprintf("failed to create test match: %v", err))
	}
	return match
}

// TestGetH verifies the logic for finding the next available height in a stick.
func TestGetH(t *testing.T) {
	match := createTestMatch3D(MatchOpts3D{R: 4, C: 4, H: 4, A: 4})
	match.Board[1][2][0] = SLOT_PLAYER1
	match.Board[1][2][1] = SLOT_PLAYER2

	if h := match.getH(0, 0); h != 0 {
		t.Errorf("Expected h=0 for empty stick (0,0), got %d", h)
	}
	if h := match.getH(1, 2); h != 2 {
		t.Errorf("Expected h=2 for partially filled stick (1,2), got %d", h)
	}

	match.Board[1][2][2] = SLOT_PLAYER1
	match.Board[1][2][3] = SLOT_PLAYER2
	if h := match.getH(1, 2); h != 4 {
		t.Errorf("Expected h=4 for full stick (1,2), got %d", h)
	}
}

// TestGetVictoryLine3D tests the victory line detection logic in all 13 3D directions.
func TestGetVictoryLine3D(t *testing.T) {
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4}

	// A map of direction names to Direction3D structs for easier test case definition.
	directions := map[string]Direction3D{
		// Axis-aligned
		"along H (vertical)": {0, 0, 1},
		"along C":            {0, 1, 0},
		"along R":            {1, 0, 0},
		// Planar diagonals
		"plane R-C diag 1": {1, 1, 0},
		"plane R-C diag 2": {1, -1, 0},
		"plane R-H diag 1": {1, 0, 1},
		"plane R-H diag 2": {1, 0, -1},
		"plane C-H diag 1": {0, 1, 1},
		"plane C-H diag 2": {0, 1, -1},
		// Space diagonals
		"space diag 1": {1, 1, 1},
		"space diag 2": {1, 1, -1},
		"space diag 3": {1, -1, 1},
		"space diag 4": {-1, 1, 1},
	}

	for name, dir := range directions {
		t.Run(name, func(t *testing.T) {
			match := createTestMatch3D(opts)
			// Place 4 pieces in a line along the given direction.
			// We start from a point and move in one direction to place pieces.
			// This avoids issues with board boundaries for any direction.
			start := Point3D{Row: 0, Col: 0, H: 0}
			// Adjust start point to ensure the line fits on the 4x4x4 board
			if dir.Row < 0 {
				start.Row = 3
			}
			if dir.Col < 0 {
				start.Col = 3
			}
			if dir.H < 0 {
				start.H = 3
			}

			points := make([]Point3D, 4)
			for i := 0; i < 4; i++ {
				p := Point3D{
					Row: start.Row + i*dir.Row,
					Col: start.Col + i*dir.Col,
					H:   start.H + i*dir.H,
				}
				// Basic boundary check for test setup
				if p.Row < 0 || p.Row >= opts.R || p.Col < 0 || p.Col >= opts.C || p.H < 0 || p.H >= opts.H {
					t.Fatalf("Test setup error: point %v is out of bounds for direction %v", p, dir)
				}
				match.Board[p.Row][p.Col][p.H] = SLOT_PLAYER1
				points[i] = p
			}

			// Check for victory from the second point in the line.
			checkPoint := points[1]
			line := match.getVictoryLine(checkPoint.Row, checkPoint.Col, checkPoint.H, dir)

			if len(line) < 4 {
				t.Errorf("Failed to detect win for direction %v. Expected line of length >= 4, got %d. Line: %v", dir, len(line), line)
			}
		})
	}

	t.Run("No Win - Interrupted Line", func(t *testing.T) {
		match := createTestMatch3D(opts)
		dir := Direction3D{Row: 1, Col: 0, H: 0}
		match.Board[0][0][0] = SLOT_PLAYER1
		match.Board[1][0][0] = SLOT_PLAYER1
		match.Board[2][0][0] = SLOT_PLAYER2 // Interruption
		match.Board[3][0][0] = SLOT_PLAYER1
		line := match.getVictoryLine(1, 0, 0, dir)
		if line != nil {
			t.Fatalf("Incorrectly found a winning line where it was interrupted: %v", line)
		}
	})
}

// TestIsGameover3D tests the top-level game over logic.
func TestIsGameover3D(t *testing.T) {
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4}
	directions := map[string]Direction3D{
		"R_axis":         {Row: 1, Col: 0, H: 0},
		"C_axis":         {Row: 0, Col: 1, H: 0},
		"H_axis":         {Row: 0, Col: 0, H: 1},
		"plane_RC_diag1": {Row: 1, Col: 1, H: 0},
		"plane_RC_diag2": {Row: 1, Col: -1, H: 0},
		"plane_RH_diag1": {Row: 1, Col: 0, H: 1},
		"plane_RH_diag2": {Row: 1, Col: 0, H: -1},
		"plane_CH_diag1": {Row: 0, Col: 1, H: 1},
		"plane_CH_diag2": {Row: 0, Col: 1, H: -1},
		"space_diag1":    {Row: 1, Col: 1, H: 1},
		"space_diag2":    {Row: 1, Col: 1, H: -1},
		"space_diag3":    {Row: 1, Col: -1, H: 1},
		"space_diag4":    {Row: -1, Col: 1, H: 1},
	}

	t.Run("Game Not Over", func(t *testing.T) {
		match := createTestMatch3D(opts)
		match.Board[0][0][0] = SLOT_PLAYER1
		match.Moves = []Move3D{{Row: 0, Col: 0}}
		if res := match.isGameover(0, 0, 0); res != nil {
			t.Errorf("Expected no gameover, but got %v", res)
		}
	})

	for name, dir := range directions {
		t.Run(fmt.Sprintf("Game Won - %s", name), func(t *testing.T) {
			match := createTestMatch3D(opts)
			start := Point3D{Row: 0, Col: 0, H: 0}
			if dir.Row < 0 {
				start.Row = 3
			}
			if dir.Col < 0 {
				start.Col = 3
			}
			if dir.H < 0 {
				start.H = 3
			}

			points := make([]Point3D, 4)
			for i := 0; i < 4; i++ {
				p := Point3D{
					Row: start.Row + i*dir.Row,
					Col: start.Col + i*dir.Col,
					H:   start.H + i*dir.H,
				}
				match.Board[p.Row][p.Col][p.H] = SLOT_PLAYER1
				points[i] = p
			}
			match.Moves = make([]Move3D, 4)

			// check from the last placed piece
			checkPoint := points[3]
			res := match.isGameover(checkPoint.Row, checkPoint.Col, checkPoint.H)
			if res == nil {
				t.Fatal("Expected a win result, but got nil")
			}
			if res["resType"] != RESULT_TYPE_WON {
				t.Errorf("Expected result type WON, but got %v", res["resType"])
			}
			lines, ok := res["lines"].([]Line3D)
			if !ok || len(lines) == 0 {
				t.Errorf("Expected non-empty lines array, got %v", res["lines"])
			}
		})
	}

	t.Run("Game Won - 2 directions", func(t *testing.T) {
		match := createTestMatch3D(opts)
		// C-axis line, without the winning piece
		match.Board[0][1][0] = SLOT_PLAYER1
		match.Board[0][2][0] = SLOT_PLAYER1
		match.Board[0][3][0] = SLOT_PLAYER1
		// R-axis line, without the winning piece
		match.Board[1][0][0] = SLOT_PLAYER1
		match.Board[2][0][0] = SLOT_PLAYER1
		match.Board[3][0][0] = SLOT_PLAYER1
		match.Moves = make([]Move3D, 7)

		// last move at (0,0,0), completes both lines
		match.Board[0][0][0] = SLOT_PLAYER1
		res := match.isGameover(0, 0, 0)
		if res == nil {
			t.Fatal("Expected a win result, but got nil")
		}
		if res["resType"] != RESULT_TYPE_WON {
			t.Errorf("Expected result type WON, but got %v", res["resType"])
		}
		lines, ok := res["lines"].([]Line3D)
		if !ok || len(lines) < 2 {
			t.Errorf("Expected at least 2 winning lines, got %d", len(lines))
		}
	})

	t.Run("Game Won - 3 directions", func(t *testing.T) {
		match := createTestMatch3D(opts)
		// C-axis line
		match.Board[0][1][0] = SLOT_PLAYER1
		match.Board[0][2][0] = SLOT_PLAYER1
		match.Board[0][3][0] = SLOT_PLAYER1
		// R-axis line
		match.Board[1][0][0] = SLOT_PLAYER1
		match.Board[2][0][0] = SLOT_PLAYER1
		match.Board[3][0][0] = SLOT_PLAYER1
		// H-axis line
		match.Board[0][0][1] = SLOT_PLAYER1
		match.Board[0][0][2] = SLOT_PLAYER1
		match.Board[0][0][3] = SLOT_PLAYER1
		match.Moves = make([]Move3D, 9)

		// last move at (0,0,0)
		match.Board[0][0][0] = SLOT_PLAYER1
		res := match.isGameover(0, 0, 0)
		if res == nil {
			t.Fatal("Expected a win result, but got nil")
		}
		if res["resType"] != RESULT_TYPE_WON {
			t.Errorf("Expected result type WON, but got %v", res["resType"])
		}
		lines, ok := res["lines"].([]Line3D)
		if !ok || len(lines) < 3 {
			t.Errorf("Expected at least 3 winning lines, got %d", len(lines))
		}
	})

	t.Run("Game is a Draw", func(t *testing.T) {
		match := createTestMatch3D(opts)
		match.Moves = make([]Move3D, opts.R*opts.C*opts.H)
		res := match.isGameover(0, 0, 0)
		if res == nil {
			t.Fatal("Expected a draw result, but got nil")
		}
		if res["resType"] != RESULT_TYPE_DRAW {
			t.Errorf("Expected result type DRAW, but got %v", res["resType"])
		}
	})
}

// TestRegisterMove3D tests the main game logic function for 3D matches.
func TestRegisterMove3D(t *testing.T) {
	p1ID := "player1"
	p2ID := "player2"
	opts := MatchOpts3D{R: 4, C: 4, H: 4, A: 4, Starts1: true}
	match := createTestMatch3D(opts)
	match.P1.ID = p1ID
	match.P2.ID = p2ID

	t.Run("Successful move", func(t *testing.T) {
		res, err := match.RegisterMove(Move3D{Row: 1, Col: 2}, p1ID)
		if err != nil {
			t.Fatalf("Expected valid move, got error: %v", err)
		}
		if res != nil {
			t.Fatalf("Expected no game over result on first move, got: %v", res)
		}
		if match.Board[1][2][0] != SLOT_PLAYER1 {
			t.Error("Board not updated correctly after P1 move")
		}
		if len(match.Moves) != 1 {
			t.Error("Move history not updated correctly")
		}
	})

	t.Run("Move by wrong player", func(t *testing.T) {
		_, err := match.RegisterMove(Move3D{Row: 0, Col: 0}, p1ID) // P1 tries to move again
		if err == nil || err.Error() != "not your turn" {
			t.Errorf("Expected 'not your turn' error, got: %v", err)
		}
	})

	t.Run("Move in invalid location", func(t *testing.T) {
		// It's P2's turn.
		_, err := match.RegisterMove(Move3D{Row: -1, Col: 0}, p2ID)
		if err == nil || err.Error() != "invalid row" {
			t.Errorf("Expected 'invalid row' error for negative row, got: %v", err)
		}
		_, err = match.RegisterMove(Move3D{Row: 4, Col: 0}, p2ID)
		if err == nil || err.Error() != "invalid row" {
			t.Errorf("Expected 'invalid row' error for out of bounds row, got: %v", err)
		}
		_, err = match.RegisterMove(Move3D{Row: 0, Col: -1}, p2ID)
		if err == nil || err.Error() != "invalid column" {
			t.Errorf("Expected 'invalid column' error for negative col, got: %v", err)
		}
		_, err = match.RegisterMove(Move3D{Row: 0, Col: 4}, p2ID)
		if err == nil || err.Error() != "invalid column" {
			t.Errorf("Expected 'invalid column' error for out of bounds col, got: %v", err)
		}
	})

	t.Run("Move in full stick", func(t *testing.T) {
		// It's P2's turn.
		_, _ = match.RegisterMove(Move3D{Row: 1, Col: 2}, p2ID) // h=1
		_, _ = match.RegisterMove(Move3D{Row: 1, Col: 2}, p1ID) // h=2
		_, _ = match.RegisterMove(Move3D{Row: 1, Col: 2}, p2ID) // h=3
		// Stick at (1,2) is now full. It's P1's turn.
		_, err := match.RegisterMove(Move3D{Row: 1, Col: 2}, p1ID)
		if err == nil || err.Error() != "invalid move. stick is full" {
			t.Errorf("Expected 'stick is full' error, got: %v", err)
		}
	})

	t.Run("Move resulting in a win", func(t *testing.T) {
		winMatch := createTestMatch3D(opts)
		winMatch.P1.ID, winMatch.P2.ID = p1ID, p2ID
		// Setup board for a win
		_, _ = winMatch.RegisterMove(Move3D{Row: 0, Col: 0}, p1ID) // h=0
		_, _ = winMatch.RegisterMove(Move3D{Row: 1, Col: 0}, p2ID) // h=0
		_, _ = winMatch.RegisterMove(Move3D{Row: 0, Col: 0}, p1ID) // h=1
		_, _ = winMatch.RegisterMove(Move3D{Row: 1, Col: 1}, p2ID) // h=0
		_, _ = winMatch.RegisterMove(Move3D{Row: 0, Col: 0}, p1ID) // h=2
		_, _ = winMatch.RegisterMove(Move3D{Row: 1, Col: 2}, p2ID) // h=0
		// Final winning move for P1 (vertical at 0,0)
		res, err := winMatch.RegisterMove(Move3D{Row: 0, Col: 0}, p1ID) // h=3

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

func (d *Direction3D) String() string {
	return fmt.Sprintf("{R:%d,C:%d,H:%d}", d.Row, d.Col, d.H)
}