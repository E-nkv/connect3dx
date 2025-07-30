package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func createJar(userID, wsURL string) (http.CookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, err
	}

	cookieURL := &url.URL{
		Scheme: "http",
		Host:   u.Host,
		Path:   "/",
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: userID, //TODO: encode this userID into jwt instead
		Path:  "/",
	}

	jar.SetCookies(cookieURL, []*http.Cookie{cookie})
	return jar, nil
}
func main() {
	wsURL := "ws://localhost:8080/ws"
	jar, err := createJar("user123", wsURL)
	if err != nil {
		fmt.Println("err creating jar", err)
		return
	}
	dialer := websocket.Dialer{
		Jar: jar,
	}

	// Dial the WebSocket URL with cookies sent automatically via the jar
	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			log.Fatalf("Dial error: %v (HTTP status code: %d)", err, resp.StatusCode)
		} else {
			log.Fatalf("Dial error: %v", err)
		}
	}
	defer conn.Close()

	fmt.Println("WebSocket connected with cookies sent")

	// Send a test message to the server
	err = conn.WriteMessage(websocket.BinaryMessage, []byte("Hello server"))
	if err != nil {
		log.Fatal("write error:", err)
	}

	// Read echo from server
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("read error:", err)
	}
	fmt.Printf("Received from server: %s\n", message)

	// Keep alive for a short demo
	time.Sleep(time.Second)
}
