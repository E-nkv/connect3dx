package api

import (
	"connectx/src/api/hub"
	"connectx/src/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type App struct {
	Hub *hub.Hub
}

func NewApp() *App {
	userModel := &models.User{}
	return &App{
		Hub: hub.NewHub(userModel),
	}
}

var upg = websocket.Upgrader{
	HandshakeTimeout:  time.Second * 3,
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
}

func (app *App) HandleWs(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		fmt.Println("token not found in cookie")
		return
	}
	fmt.Println("token is: ", tokenCookie.Value)
	conn, err := upg.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("err upgrading conn: ", conn)
		return
	}

	go app.Hub.ListenFromUser(tokenCookie.Value, conn)
}
