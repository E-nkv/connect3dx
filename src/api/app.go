package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type App struct {
	Hub *Hub
}

func NewApp() *App {
	return &App{
		Hub: newHub(),
	}
}

var upg = websocket.Upgrader{
	HandshakeTimeout:  time.Second * 3,
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
}

func (app *App) HandleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upg.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("err upgrading conn: ", conn)
		return
	}
	/* userID := r.URL.Query().Get("id")
	if userID == "" {
		panic("userid empty")
	} */
	go app.Hub.ListenFromUser("testID", conn)
}
