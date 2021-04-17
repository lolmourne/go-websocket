// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	authClient "github.com/lolmourne/go-accounts/client/userauth"
	"github.com/lolmourne/go-groupchat/client"
)

var addr = flag.String("addr", ":90", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()

	chString := make(chan string)
	go func(ch chan string) {
		for {
			select {
			case msg := <-ch:
				log.Println(msg, "from channel")
			}
		}

	}(chString)

	gcClient := client.NewClient("http://localhost:8080", time.Duration(30)*time.Second)
	auCli := authClient.NewClient("http://localhost:7070", time.Duration(30)*time.Second)

	roomMgr := NewRoomManager(gcClient)
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ch_test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Write([]byte("method unallowed"))
			return
		}

		msg := r.FormValue("message")
		chString <- msg
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(roomMgr, auCli, w, r)
	})
	log.Println("RUNNING----")
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
