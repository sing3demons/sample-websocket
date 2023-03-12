package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ChatServer struct {
	clients map[*websocket.Conn]struct{}
	mu      sync.Mutex
}

func (cs *ChatServer) AddClient(conn *websocket.Conn) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.clients[conn] = struct{}{}
}

func (cs *ChatServer) RemoveClient(conn *websocket.Conn) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.clients, conn)
}

func (cs *ChatServer) Broadcast(msg []byte) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for conn := range cs.clients {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func main() {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	chatServer := &ChatServer{
		clients: make(map[*websocket.Conn]struct{}),
		mu:      sync.Mutex{},
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		defer conn.Close()

		chatServer.AddClient(conn)
		defer chatServer.RemoveClient(conn)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			log.Println(string(msg))
			chatServer.Broadcast(msg)
			// err = conn.WriteMessage(websocket.TextMessage, msg)
			// if err != nil {
			// 	log.Println(err)
			// 	return
			// }
		}

	})

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
