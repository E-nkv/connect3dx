package types

import (
	"time"
)

type RegisterMovePL struct {
	MatchID string    `json:"match_id"`
	Col     int       `json:"col"`
	SentAt  time.Time `json:"sent_at"`
}

type JoinMatchPL struct {
	MatchID string `json:"match_id"`
}
