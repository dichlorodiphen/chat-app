// A minimal echo server.
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	hub := newHub()
	go hub.run()

	http.HandleFunc("/", echo)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	fmt.Println("Starting server.")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// Echoes back the request path to the client.
func echo(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	fmt.Println("Got connection")
	fmt.Fprintf(w, "Path: %q\n", r.URL.Path)
}
