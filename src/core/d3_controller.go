package core

import (
	"connectx/src/errs"
	"connectx/src/types"
	"strings"
	"sync"
	"time"

	"fmt"

	"github.com/google/uuid"
)

type MatchController3D struct {
	Matches      map[string]*Match3D
	MatchesMutex sync.Mutex
}

func NewMatchController3D() *MatchController3D {
	return &MatchController3D{
		Matches: make(map[string]*Match3D),
	}
}

func (c *MatchController3D) CreateMatch(p1ID string, opts MatchOpts3D) (string, error) {
	if err := validMatchOptions3D(opts); err != nil {
		return "", fmt.Errorf("invalid match options: %s", err.Error())
	}
	m, err := NewMatch3D(p1ID, "", opts)
	if err != nil {
		return "", err
	}
	id := uuid.New().String()
	c.MatchesMutex.Lock()
	c.Matches[id] = m
	c.MatchesMutex.Unlock()
	return id, nil
}

func (c *MatchController3D) JoinMatch(playerID string, matchID string) (*Match3D, bool, error) {
	c.MatchesMutex.Lock()
	defer c.MatchesMutex.Unlock()
	match, ok := c.Matches[matchID]
	if !ok {
		return nil, false, errs.ErrNotFound
	}
	if match.P2.ID != "" {
		if match.P1.ID != playerID && match.P2.ID != playerID {
			return nil, false, errs.ErrUnjoinable
		}
		return match, false, nil
	}

	//first time that user2 joins
	match.P2.ID = playerID
	match.Started = true
	match.StartedAt = time.Now()
	return match, true, nil
}

func validMatchOptions3D(opts MatchOpts3D) error {
	errs := []string{}
	if opts.R < 3 || opts.R > 10 {
		errs = append(errs, "invalid R")
	}
	if opts.C < 3 || opts.C > 10 {
		errs = append(errs, "invalid C")
	}
	if opts.H < 3 || opts.H > 10 {
		errs = append(errs, "invalid H")
	}
	if opts.A < 3 || opts.A > 10 {
		errs = append(errs, "invalid A")
	}
	if opts.A > opts.R || opts.A > opts.C || opts.A > opts.H {
		errs = append(errs, "A cant be bigger than R, C, nor H")
	}
	errStr := strings.Join(errs, ", ")
	if errStr != "" {
		return fmt.Errorf("%s", errStr)
	}
	return nil

}

func (c *MatchController3D) RegisterMove(userID string, pl types.RegisterMove3DPL) (*Match3D, GameoverResult3D, error) {
	c.MatchesMutex.Lock()
	m, ok := c.Matches[pl.MatchID]
	c.MatchesMutex.Unlock()
	if !ok {
		return nil, nil, errs.ErrNotFound
	}
	move := Move3D{Col: pl.Col, Row: pl.Row, RegisteredAt: time.Now()}
	res, err := m.RegisterMove(move, userID)
	if err != nil {
		return nil, nil, err
	}
	return m, res, nil
}
