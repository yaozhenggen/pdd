package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

type PddAuth struct {
	Request_id  string
	Response    string
	Result      string
	Server_time string
	Status      string
	Uid         string
}

type PddChat struct {
	Message    Message
	Request_id string
	Response   string
	Result     string
}

type Message struct {
	Content string
	Ts      string
	From    User
	To      User
	Info    Info
}

type Info struct {
	CustomerNumber int
	GoodsID        string
	GoodsName      string
	GoodsPrice     string
	GoodsThumbUrl  string
	Url            string
	ts             string
}

type SendMessage struct {
	Cmd        string
	Request_id int
	Message    Message
	Random     string
}

type User struct {
	Role string
	Uid  string
}

type Chat_Log struct {
	Id        int
	Text      string
	Time      time.Time
	Token     string
	User      string
	To_user   string
	Form_user string
}

func pddWs(client *Client) {
	pddConn := client.request
	u := url.URL{Scheme: "wss", Host: *pddaddr, Path: "/"}
	q := u.Query()
	q.Set("access_token", pddConn["access_token"])
	q.Set("role", pddConn["role"])
	q.Set("client", pddConn["client"])
	q.Set("version", pddConn["version"])
	u.RawQuery = q.Encode()

	pddClient, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return
	}

	// init a Client
	pdd := &Client{hub: pddHub, send: make(chan []byte, 256), ws: pddClient, request: pddConn}
	pdd.hub.register <- pdd
	relationClient[pddConn["access_token"]] = []*Client{client, pdd}

	go pdd.writePump()
	go pdd.readPump()
}



func addChat(data *Chat_Log) error {
	stmt, err := engine.Prepare(`INSERT chat_log (text,time,token,user,to_user,from_user) VALUES (?,?,?,?,?,?)`)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(data.Text, data.Time, data.Token, data.User, data.To_user, data.Form_user)
	if err != nil {
		return err
	}

	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}
