package core

import (
	"fmt"
	"time"
)

type Move3D struct {
	Col          int
	Row          int
	RegisteredAt time.Time
}
type MatchOpts3D struct {
	R       int   `json:"r"`
	C       int   `json:"c"`
	H       int   `json:"h"`
	A       int   `json:"a"`
	Starts1 bool  `json:"starts1"`
	T0      int64 `json:"t0"`
	TD      int64 `json:"td"`
}

type Point3D struct {
	Row int
	Col int
	H   int
}
type Line3D []Point3D

func createBoard3D(R, C, H int) [][][]Slot {
	board := make([][][]Slot, R)
	for i := 0; i < R; i++ {
		board[i] = make([][]Slot, C)
		for j := 0; j < C; j++ {
			board[i][j] = make([]Slot, H)
		}
	}
	return board
}

func NewMatch3D(p1ID, p2ID string, opts MatchOpts3D) (*Match3D, error) {
	if opts.R <= 0 || opts.C <= 0 || opts.H <= 0 || opts.A <= 0 {
		return nil, fmt.Errorf("invalid match options: dimensions and alignment must be positive")
	}
	if opts.A > opts.R && opts.A > opts.C && opts.A > opts.H {
		return nil, fmt.Errorf("invalid match options: alignment must be less than or equal to any dimension")
	}
	return &Match3D{
		Opts: opts,
		P1: Player{
			ID:       p1ID,
			TimeLeft: opts.T0,
		},
		P2: Player{
			ID:       p2ID,
			TimeLeft: opts.T0,
		},
		Board: createBoard3D(opts.R, opts.C, opts.H),
		Moves: make([]Move3D, 0),
	}, nil
}

type Match3D struct {
	Board     [][][]Slot
	P1        Player
	P2        Player
	Opts      MatchOpts3D
	Moves     []Move3D
	StartedAt time.Time
	Started   bool
	Gameover  bool
}

type Match3DDTO struct {
	Board     [][][]int `json:"board"`
	P1        *PlayerDTO
	P2        *PlayerDTO
	Opts      MatchOpts3D
	Moves     []Move3D
	StartedAt time.Time
	Started   bool
	Gameover  bool
}

func (m *Match3D) ToDTO(userModel DTOGetter) (*Match3DDTO, error) {
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
	boardDTO := make([][][]int, len(m.Board))
	for i, layer := range m.Board {
		boardDTO[i] = make([][]int, len(layer))
		for j, row := range layer {
			boardDTO[i][j] = make([]int, len(row))
			for k, slot := range row {
				boardDTO[i][j][k] = int(slot)
			}
		}
	}

	return &Match3DDTO{
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


type GameoverResult3D map[string]any

func (m *Match3D) getCurrPlayerID() string {
	if (len(m.Moves)%2 == 0) == m.Opts.Starts1 {
		return m.P1.ID
	}
	return m.P2.ID
}

func (m *Match3D) getH(row, col int) int {
	stick := m.Board[row][col]
	for h, val := range stick {
		if val == SLOT_EMPTY {
			return h
		}
	}
	return m.Opts.H
}

func (m *Match3D) GetEnemyID(pid string) string {
	if pid != m.P1.ID && pid != m.P2.ID {
		return ""
	}
	if pid == m.P1.ID {
		return m.P2.ID
	}
	return m.P1.ID
}

type Direction3D struct {
	Row int
	Col int
	H   int
}

func (d *Direction3D) OtherSide() Direction3D {
	r := 0
	if d.Row != 0 {
		r = -d.Row
	}
	c := 0
	if d.Col != 0 {
		c = -d.Col
	}
	h := 0
	if d.H != 0 {
		h = -d.H
	}
	return Direction3D{Row: r, Col: c, H: h}
}

func (m *Match3D) getVictoryLine(row, col, h int, dir Direction3D) Line3D {
	v := m.Board[row][col][h]
	if v == SLOT_EMPTY {
		return nil
	}
	line := Line3D{Point3D{Row: row, Col: col, H: h}}

	// Check in the given direction
	for i, j, k := row+dir.Row, col+dir.Col, h+dir.H; i >= 0 && i < m.Opts.R && j >= 0 && j < m.Opts.C && k >= 0 && k < m.Opts.H; i, j, k = i+dir.Row, j+dir.Col, k+dir.H {
		if m.Board[i][j][k] == v {
			line = append(line, Point3D{Row: i, Col: j, H: k})
		} else {
			break
		}
	}

	// Check in the other direction
	otherDir := dir.OtherSide()
	for i, j, k := row+otherDir.Row, col+otherDir.Col, h+otherDir.H; i >= 0 && i < m.Opts.R && j >= 0 && j < m.Opts.C && k >= 0 && k < m.Opts.H; i, j, k = i+otherDir.Row, j+otherDir.Col, k+otherDir.H {
		if m.Board[i][j][k] == v {
			line = append(line, Point3D{Row: i, Col: j, H: k})
		} else {
			break
		}
	}

	if len(line) >= m.Opts.A {
		return line
	}
	return nil
}
func (m *Match3D) isGameover(row, col, h int) GameoverResult3D {
	dirs := []Direction3D{
		// Axis-aligned
		{Row: 1, Col: 0, H: 0},
		{Row: 0, Col: 1, H: 0},
		{Row: 0, Col: 0, H: 1},
		// Planar diagonals
		{Row: 1, Col: 1, H: 0},
		{Row: 1, Col: -1, H: 0},
		{Row: 1, Col: 0, H: 1},
		{Row: 1, Col: 0, H: -1},
		{Row: 0, Col: 1, H: 1},
		{Row: 0, Col: 1, H: -1},
		// Space diagonals
		{Row: 1, Col: 1, H: 1},
		{Row: 1, Col: 1, H: -1},
		{Row: 1, Col: -1, H: 1},
		{Row: -1, Col: 1, H: 1},
	}

	var lines []Line3D
	for _, dir := range dirs {
		line := m.getVictoryLine(row, col, h, dir)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if len(lines) > 0 {
		return GameoverResult3D{"resType": RESULT_TYPE_WON, "lines": lines}
	}

	if len(m.Moves) >= m.Opts.R*m.Opts.H*m.Opts.C {
		return GameoverResult3D{
			"resType": RESULT_TYPE_DRAW,
		}
	}
	return nil
}

func (m *Match3D) RegisterMove(move Move3D, pid string) (GameoverResult3D, error) {
	if m.Gameover {
		return nil, fmt.Errorf("game is over")
	}
	if !m.Started {
		return nil, fmt.Errorf("match has not started yet")
	}
	if move.Row < 0 || move.Row >= m.Opts.R {
		return nil, fmt.Errorf("invalid row")
	}
	if move.Col < 0 || move.Col >= m.Opts.C {
		return nil, fmt.Errorf("invalid column")
	}
	currPID := m.getCurrPlayerID()
	if currPID != pid {
		return nil, fmt.Errorf("not your turn")
	}
	h := m.getH(move.Row, move.Col)
	if h >= m.Opts.H {
		return nil, fmt.Errorf("invalid move. stick is full")
	}
	m.Board[move.Row][move.Col][h] = SLOT_PLAYER1
	if currPID == m.P2.ID {
		m.Board[move.Row][move.Col][h] = SLOT_PLAYER2
	}
	m.Moves = append(m.Moves, move)
	res := m.isGameover(move.Row, move.Col, h)
	if res != nil {
		m.Gameover = true
	}
	return res, nil
}