package models

import (
	"connectx/src/core"
)

type User struct {
}

func (userModel *User) GetUserDTO(userID string) (*core.PlayerDTO, error) {
	return &core.PlayerDTO{
		ID: userID,
	}, nil
}
