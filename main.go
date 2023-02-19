package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) HandleWS(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client: ", ws.RemoteAddr())

	s.conns[ws] = true
}

func (s *Server) ReadLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err)

			continue
		}

		msg := buf[:n]

		s.Broadcast(msg)
	}
}

func (s *Server) HandleWSOrderbook(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client to orderbook feed: ", ws.RemoteAddr())

	for {
		payload := fmt.Sprintf("orderbook data -> %d\n", time.Now().UnixNano())
		ws.Write([]byte(payload))
		time.Sleep(2 * time.Second)
	}
}

func (s *Server) Broadcast(b []byte) {
	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			_, err := ws.Write(b)
			if err != nil {
				fmt.Printf("write error: %v", err)
			}
		}(ws)
	}

}

func main() {
	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.HandleWS))
	http.Handle("/orderbookfeed", websocket.Handler(server.HandleWSOrderbook))
	http.ListenAndServe(":3000", nil)
}
