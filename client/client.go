// client.go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	url := "ws://localhost:8080/ws"
	fmt.Println("Dialing", url)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	for i := 0; i < 3; i++ {
		msg := fmt.Sprintf("Hello %d", i+1)
		fmt.Println("Sending:", msg)
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte(msg)); err != nil {
			log.Fatal("write error:", err)
		}

		_, reply, err := conn.ReadMessage()
		if err != nil {
			log.Fatal("read error:", err)
		}
		fmt.Println("Received from server:", string(reply))
		time.Sleep(time.Second)
	}
}
