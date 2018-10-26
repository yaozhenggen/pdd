// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"./tool"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	//0为客户端，1为拼多多
	tag int
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		tag:        0,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			if h.tag == 0 {
				go pddWs(client)
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				if h.tag == 0 {
					pddHub.unregister <- relationClient["access_token"][1]
					delete(relationClient, client.request["access_token"])
				}
			}

		case client := <-h.broadcast:
			msg := string(client.message)
			fmt.Println(msg)
			switch h.tag {
			case 0:
				fmt.Println(len(relationClient))
				pddClient := relationClient[client.request["access_token"]][1]

				if strings.Contains(msg, `"cmd":"send_message"`) {
					send := SendMessage{}
					err := json.Unmarshal(client.message, &send)
					if err != nil {
						tool.Writelog("解析聊天数据错误：" + err.Error())
						continue
					}

					data := Chat_Log{}
					data.Text = send.Message.Content
					data.Form_user = send.Message.From.Uid
					data.To_user = send.Message.To.Uid

					ts, err := strconv.ParseInt(send.Message.Ts, 10, 64)
					if err != nil {
						tool.Writelog("聊天数据入库错误：" + err.Error())
						continue
					}

					data.Time = time.Unix(ts, 0)
					data.Token = client.request["access_token"]

					err = addChat(&data)
					if err != nil {
						tool.Writelog("聊天数据入库错误：" + err.Error())
						continue
					}

					select {
					case pddClient.send <- client.message:
					default:
						h.unregister <- client
						pddHub.unregister <- pddClient
					}
				}
			case 1:
				fmt.Println(len(relationClient))
				userClient := relationClient[client.request["access_token"]][0]

				if strings.Contains(msg, `"type":30`) {
					h.unregister <- client
					clientHub.register <- userClient
				} else if ((strings.Contains(msg, `"type":0,`) || strings.Contains(msg, `"type":1,`)) && !strings.Contains(msg, `"role":"mall_cs"`)) {
					var chat PddChat
					err := json.Unmarshal(client.message, &chat)
					if err != nil {
						tool.Writelog("解析聊天数据错误：" + err.Error())
						continue
					}

					data := Chat_Log{}
					if chat.Message.Info.GoodsID != "" {
						chat.Message.Info.Url = chat.Message.Content
						info, err := json.Marshal(chat.Message.Info)
						if err != nil {
							tool.Writelog("解析聊天内容错误：" + err.Error())
							continue
						}
						data.Text = string(info)
					} else {
						data.Text = chat.Message.Content
					}

					data.Form_user = chat.Message.From.Uid
					data.To_user = chat.Message.To.Uid
					ts, err := strconv.ParseInt(chat.Message.Ts, 10, 64)
					if err != nil {
						tool.Writelog("聊天数据入库错误：" + err.Error())
						continue
					}

					data.Time = time.Unix(ts, 0)
					data.Token = client.request["access_token"]

					err = addChat(&data)
					if err != nil {
						tool.Writelog("聊天数据入库错误：" + err.Error())
						continue
					}

					message, err := json.Marshal(data)
					if err != nil {
						tool.Writelog("广播数据错误：" + err.Error())
						continue
					}

					select {
					case userClient.send <- message:
					default:
						h.unregister <- client
						clientHub.unregister <- userClient
					}
				}
			}
		}
	}
}
