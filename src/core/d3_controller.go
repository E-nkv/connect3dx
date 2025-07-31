package core

type MatchController3D struct {
	Matches map[string]*Match3D
}

func NewMatchController3D() *MatchController3D {
	return &MatchController3D{
		Matches: make(map[string]*Match3D),
	}
}
