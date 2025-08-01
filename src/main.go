package main

import (
	"connectx/src/api"
	"fmt"
	"log"
	"net/http"
)

func main() {
	app := api.NewApp()
	m := http.NewServeMux()
	m.HandleFunc("/ws", app.HandleWs)
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hi")
		fmt.Fprint(w, "hi")
	})
	fmt.Println("running at 8080")
	log.Fatal(http.ListenAndServe(":8080", m))
}
