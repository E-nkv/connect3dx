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

type Move struct {
	Col          int
	RegisteredAt time.Time
}
type MatchOpts struct {
	W       int   `json:"w"`
	H       int   `json:"h"`
	A       int   `json:"a"`
	Starts1 bool  `json:"starts1"`
	T0      int64 `json:"t0"`
	TD      int64 `json:"td"`
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

func createBoard2D(W, H int) [][]Slot {
	board := make([][]Slot, H)
	for i := 0; i < H; i++ {
		board[i] = make([]Slot, W)
	}
	return board
}

func NewMatch2D(p1ID, p2ID string, opts MatchOpts) (*Match2D, error) {
	if opts.W <= 0 || opts.H <= 0 || opts.A <= 0 {
		return nil, fmt.Errorf("invalid match options: dimensions and alignment must be positive")
	}
	if opts.A > opts.W && opts.A > opts.H {
		return nil, fmt.Errorf("invalid match options: alignment must be less than or equal to width or height")
	}
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
	}, nil
}

type Match2D struct {
	Board     [][]Slot
	P1        Player
	P2        Player
	Opts      MatchOpts
	Moves     []Move
	StartedAt time.Time
	Started   bool
	Gameover  bool
}

type Match2DDTO struct {
	Board     [][]int `json:"board"`
	P1        *PlayerDTO
	P2        *PlayerDTO
	Opts      MatchOpts
	Moves     []Move
	StartedAt time.Time
	Started   bool
	Gameover  bool
}

func (m *Match2D) ToDTO(userModel DTOGetter) (*Match2DDTO, error) {
	p1, err := userModel.GetUserDTO(m.P1.ID)
	if err != nil {
		return nil, err
	}
	p1.TimeLeft = m.P1.TimeLeft

	var p2 *PlayerDTO
	if m.P2.ID != "" {
		var err error
		p2, err = userModel.GetUserDTO(m.P2.ID)
		if err != nil {
			return nil, err
		}
		p2.TimeLeft = m.P2.TimeLeft
	}

	// Convert board
	boardDTO := make([][]int, len(m.Board))
	for i, row := range m.Board {
		boardDTO[i] = make([]int, len(row))
		for j, slot := range row {
			boardDTO[i][j] = int(slot)
		}
	}

	return &Match2DDTO{
		Board:     boardDTO,
		P1:        p1,
		P2:        p2,
		Opts:      m.Opts,
		Moves:     m.Moves,
		StartedAt: m.StartedAt,
		Started:   m.Started,
		Gameover:  m.Gameover,
	}, nil
}


type GameoverResult map[string]any

func (m *Match2D) getCurrPlayerID() string {
	if (len(m.Moves)%2 == 0) == m.Opts.Starts1 {
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

func (m *Match2D) GetEnemyID(pid string) string {
	if pid != m.P1.ID && pid != m.P2.ID {
		return ""
	}
	if pid == m.P1.ID {
		return m.P2.ID
	}
	return m.P1.ID
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
	if m.Gameover {
		return nil, fmt.Errorf("game is over")
	}
	if !m.Started {
		return nil, fmt.Errorf("match has not started yet")
	}
	if move.Col < 0 || move.Col >= m.Opts.W {
		return nil, fmt.Errorf("invalid column")
	}
	currPID := m.getCurrPlayerID()
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
	if res != nil {
		m.Gameover = true
	}
	return res, nil
}
