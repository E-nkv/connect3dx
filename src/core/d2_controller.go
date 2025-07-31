package core

import (
	"connectx/src/errs"
	"fmt"
	"strconv"
	"strings"
)

type MatchController2D struct {
	Matches map[string]*Match2D
}

func NewMatchController2D() *MatchController2D {
	return &MatchController2D{
		Matches: make(map[string]*Match2D),
	}
}

func (c *MatchController2D) CreateMatch(p1ID string, opts MatchOpts) (string, error) {
	if err := validMatchOptions(opts); err != nil {
		return "", fmt.Errorf("invalid match options: %s", err.Error())
	}
	m := NewMatch2D(p1ID, "", opts)
	id := GENERATE_UUID()
	c.Matches[id] = m
	return id, nil
}

func (c *MatchController2D) JoinMatch(playerID string, matchID string) (*Match2D, bool, error) {

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
	return match, true, nil
}

var CURRNUM = 0

func GENERATE_UUID() string {
	CURRNUM++
	return strconv.Itoa(CURRNUM)
}
func validMatchOptions(opts MatchOpts) error {
	errs := []string{}
	if opts.W < 3 || opts.W > 15 {
		errs = append(errs, "invalid W")
	}
	if opts.H < 3 || opts.H > 15 {
		errs = append(errs, "invalid H")
	}
	if opts.A < 3 || opts.A > 15 {
		errs = append(errs, "invalid A")
	}
	if opts.A > opts.W || opts.A > opts.H {
		errs = append(errs, "A cant be bigger than W nor H")
	}
	errStr := strings.Join(errs, ", ")
	if errStr != "" {
		return fmt.Errorf("%s", errStr)
	}
	return nil

}
