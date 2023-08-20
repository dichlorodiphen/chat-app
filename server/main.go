// A minimal echo server.
package main

import "log"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	s := newServer()
	s.setUpRoutes()
	s.start()
}
