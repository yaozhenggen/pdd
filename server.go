package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
)

var addr = flag.String("addrs", ":8888", "service address")
var pdd = flag.String("pdd", "m-ws.pinduoduo.com", "http service address")

func main() {
	//go tool.StartFlashServ()
	go h.run()
	println("Start Serv...")

	http.HandleFunc("/zkz", ServeWs)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte
	// hq request string
	request map[string]string
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true

			u := url.URL{Scheme: "wss", Host: *pdd, Path: "/"}
			q := u.Query()
			q.Set("access_token", c.request["access_token"])
			q.Set("role", c.request["role"])
			q.Set("client", c.request["client"])
			q.Set("version", c.request["version"])
			u.RawQuery = q.Encode()

			fmt.Println(u.String())
			client, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Fatal("dial:", err)
			}
			defer client.Close()

			done := make(chan struct{})

			go func() {
				defer close(done)
				for {
					_, message, err := client.ReadMessage()
					if err != nil {
						log.Println("read:", err)
					}
					log.Printf("recv: %s", message)
				}
			}()
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
				//log.Println("remove user", c.ws.RemoteAddr().String())
			}
		case msg := <-h.broadcast:
			log.Println("remove user", msg)
		}
	}
}

