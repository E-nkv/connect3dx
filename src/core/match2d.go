package core

import (
	"fmt"
	"time"
)

type Slot int
type RESULT_TYPE int

const (
	RESULT_TYPE_WON RESULT_TYPE = iota
	RESULT_TYPE_DRAW
	RESULT_TYPE_TIMEOUT
)
const (
	SLOT_EMPTY Slot = iota
	SLOT_PLAYER1
	SLOT_PLAYER2
)

type Direction struct {
	Row int
	Col int
}

func (d *Direction) OtherSide() Direction {
	r := 0
	if d.Row != 0 {
		r = -d.Row
	}
	c := 0
	if d.Col != 0 {
		c = -d.Col
	}
	return Direction{Row: r, Col: c}

}

type Point struct {
	Row int
	Col int
}
type Line []Point
type Player struct {
	ID       string
	TimeLeft int64
}

type MatchOpts struct {
	W       int   `json:"w"`
	H       int   `json:"h"`
	A       int   `json:"a"`
	Starts1 bool  `json:"starts1"`
	T0      int64 `json:"t0"` //Initial time in seconds
	TD      int64 `json:"td"` //Seconds to add per move
}

// FIX: The loop `for i := range H` was incorrect. It should be a standard
// for loop `for i := 0; i < H; i++` to initialize the board rows.
func createBoard2D(W, H int) [][]Slot {
	board := make([][]Slot, H)
	for i := 0; i < H; i++ {
		board[i] = make([]Slot, W)
	}
	return board
}

func NewMatch2D(p1ID, p2ID string, opts MatchOpts) *Match2D {
	return &Match2D{
		Opts: opts,
		P1: Player{
			ID:       p1ID,
			TimeLeft: opts.T0,
		},
		P2: Player{
			ID:       p2ID,
			TimeLeft: opts.T0,
		},
		Board: createBoard2D(opts.W, opts.H),
		Moves: make([]Move, 0),
	}
}

type Move struct {
	Col          int
	RegisteredAt time.Time
}

type Match2D struct {
	Board     [][]Slot
	P1        Player
	P2        Player
	Opts      MatchOpts
	Moves     []Move
	StartedAt time.Time
}

type GameoverResult map[string]any

func (m *Match2D) getCurrPlayer() string {
	moves1 := false
	if (m.Opts.Starts1 && len(m.Moves)%2 == 0) || (!m.Opts.Starts1 && len(m.Moves)%2 == 1) {
		moves1 = true
	}
	if moves1 {
		return m.P1.ID
	}
	return m.P2.ID
}
func (m *Match2D) getRow(col int) int {
	for i := m.Opts.H - 1; i >= 0; i-- {
		if m.Board[i][col] == SLOT_EMPTY {
			return i
		}
	}
	return -1
}

// FIX: The second for-loop was starting from the wrong position. It should start
// from the original point (row, col) and iterate in the `otherDir` direction.
// The original code was restarting from `row+dir.Row`, which double-counted pieces.
func (m *Match2D) getVictoryLine(row, col int, dir Direction) Line {
	v := m.Board[row][col]
	if v == SLOT_EMPTY {
		return nil
	}
	line := Line{Point{Row: row, Col: col}}

	// Check in the primary direction
	for i, j := row+dir.Row, col+dir.Col; i >= 0 && i < m.Opts.H && j >= 0 && j < m.Opts.W && m.Board[i][j] == v; i, j = i+dir.Row, j+dir.Col {
		line = append(line, Point{Row: i, Col: j})
	}

	// Check in the opposite direction
	otherDir := dir.OtherSide()
	for i, j := row+otherDir.Row, col+otherDir.Col; i >= 0 && i < m.Opts.H && j >= 0 && j < m.Opts.W && m.Board[i][j] == v; i, j = i+otherDir.Row, j+otherDir.Col {
		line = append(line, Point{Row: i, Col: j})
	}

	if len(line) >= m.Opts.A {
		return line
	}
	return nil
}

// FIX: The check for a win was `if lines != nil`, which is always true for an
// initialized slice. It must be changed to `if len(lines) > 0`.
// The logic is also reordered to check for a win first, then a draw.
func (m *Match2D) isGameover(row, col int) GameoverResult {
	dirs := []Direction{
		{Row: 1, Col: 0},  // Vertical
		{Row: 0, Col: 1},  // Horizontal
		{Row: 1, Col: 1},  // Diagonal
		{Row: 1, Col: -1}, // Anti-Diagonal (Corrected from previous d/ad mixup)
	}
	var lines []Line
	for _, dir := range dirs {
		line := m.getVictoryLine(row, col, dir)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// If any winning lines were found, the game is won.
	if len(lines) > 0 {
		return GameoverResult{"resType": RESULT_TYPE_WON, "lines": lines}
	}

	// If no win, check if the board is full (draw).
	if len(m.Moves) == m.Opts.H*m.Opts.W {
		return GameoverResult{"resType": RESULT_TYPE_DRAW}
	}

	// No win, no draw, game continues.
	return nil
}

// returns whether it was a normal move, the result if it was a decisive move, and the error if any
func (m *Match2D) RegisterMove(move Move, pid string) (GameoverResult, error) {
	if move.Col < 0 || move.Col >= m.Opts.W {
		return nil, fmt.Errorf("invalid column")
	}
	currPID := m.getCurrPlayer()
	if currPID != pid {
		return nil, fmt.Errorf("not your turn")
	}
	row := m.getRow(move.Col)
	if row <= -1 {
		return nil, fmt.Errorf("invalid move. column is full")
	}
	m.Board[row][move.Col] = SLOT_PLAYER1
	if currPID == m.P2.ID {
		m.Board[row][move.Col] = SLOT_PLAYER2
	}
	m.Moves = append(m.Moves, move)
	res := m.isGameover(row, move.Col)
	return res, nil
}
