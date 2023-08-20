// A minimal echo server.
package main

func main() {
	s := newServer()
	s.setUpRoutes()
	s.start()
}
