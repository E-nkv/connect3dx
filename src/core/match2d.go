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
	T0      int64 `json:"t0"`
	TD      int64 `json:"td"`
}

func createBoard2D(W, H int) [][]Slot {
	board := make([][]Slot, H)
	for i := range H {
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

func (m *Match2D) getVictoryLine(row, col int, dir Direction) Line {
	v := m.Board[row][col]
	if v == SLOT_EMPTY {
		return nil
	}
	line := Line{Point{Row: row, Col: col}}

	for i, j := row+dir.Row, col+dir.Col; i >= 0 && i < m.Opts.H && j >= 0 && j < m.Opts.W && m.Board[i][j] == v; i, j = i+dir.Row, j+dir.Col {
		line = append(line, Point{Row: i, Col: j})
	}

	otherDir := dir.OtherSide()
	for i, j := row+otherDir.Row, col+otherDir.Col; i >= 0 && i < m.Opts.H && j >= 0 && j < m.Opts.W && m.Board[i][j] == v; i, j = i+otherDir.Row, j+otherDir.Col {
		line = append(line, Point{Row: i, Col: j})
	}

	if len(line) >= m.Opts.A {
		return line
	}
	return nil
}

func (m *Match2D) isGameover(row, col int) GameoverResult {
	dirs := []Direction{
		{Row: 1, Col: 0},
		{Row: 0, Col: 1},
		{Row: 1, Col: 1},
		{Row: 1, Col: -1},
	}
	var lines []Line
	for _, dir := range dirs {
		line := m.getVictoryLine(row, col, dir)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if len(lines) > 0 {
		return GameoverResult{"resType": RESULT_TYPE_WON, "lines": lines}
	}

	if len(m.Moves) == m.Opts.H*m.Opts.W {
		return GameoverResult{"resType": RESULT_TYPE_DRAW}
	}

	return nil
}

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
