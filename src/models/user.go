package models

import (
	"connectx/src/api/hub"
)

type User struct {
}

func (userModel *User) GetUserDTO(userID string) (*hub.PlayerDTO, error) {

}
