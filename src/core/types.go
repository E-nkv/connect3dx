package core

type PlayerDTO struct {
	ID       string `json:"id"`
	TimeLeft int64  `json:"timeLeft"`
	Nick     string `json:"nick"`
	ImgURL   string `json:"imgUrl"`
}

type DTOGetter interface {
	GetUserDTO(userID string) (*PlayerDTO, error)
}
