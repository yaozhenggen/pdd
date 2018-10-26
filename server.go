package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"flag"
	"log"
	"net/http"
	"./tool"
)

var dbInfo = []string{"localhost:3306", "root", "Mysql123", "zkz"}
var addr = flag.String("addrs", ":8888", "service address")
var pddaddr = flag.String("pdd", "m-ws.pinduoduo.com", "http service address")
var pddHub = newHub()
var clientHub = newHub()
var engine *sql.DB

func main() {
	//go tool.StartFlashServ()
	clientHub.tag = 0
	go clientHub.run()

	pddHub.tag = 1
	go pddHub.run()

	var err error
	engine, err = sql.Open("mysql", dbInfo[1]+":"+dbInfo[2]+"@tcp("+dbInfo[0]+")/"+dbInfo[3]+"?charset=utf8&parseTime=true&loc=Local")
	if err != nil {
		log.Fatal(err)
		tool.Writelog(err.Error())
	}

	println("Start Serv...")
	http.HandleFunc("/zkz", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}


}
