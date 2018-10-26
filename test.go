// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var addr = flag.String("addr", "127.0.0.1:8888", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	//access_token=5f95cd4847a53b326d9d8c447c5ce885&role=mall_cs&client=web&version=201810172014

//wss://m-ws.pinduoduo.com/?access_token=f5f8dd26eabc134bb0194eee47031640&role=mall_cs&client=web&version=201810241339

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/zkz"}
	q := u.Query()
	q.Set("access_token", "78bdbb9b6ec646270797cb190d302f3e")
	//q.Set("access_token", "f5f8dd26eabc134bb0194eee47031640")
	q.Set("role", "mall_cs")
	q.Set("client", "web")
	q.Set("version", "201810172014")
	u.RawQuery = q.Encode()

	//u, err := url.Parse("http://bing.com/search?q=dotnet")
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
