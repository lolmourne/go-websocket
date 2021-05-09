// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	authClient "github.com/lolmourne/go-accounts/client/userauth"
	"github.com/lolmourne/go-groupchat/client"
	"github.com/lolmourne/go-websocket/resource/chat"
	"github.com/lolmourne/go-websocket/resource/user"
	redisCli "github.com/lolmourne/r-pipeline/client"
	"github.com/lolmourne/r-pipeline/pubsub"
)

var addr = flag.String("addr", ":90", "http service address")
var sub pubsub.RedisPubsub
var rdb *redis.Client

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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	rdb = redis.NewClient(&redis.Options{
		Addr:     "34.101.216.10:6379",
		Password: "skilvulredis", // no password set
		DB:       0,              // use default DB
	})

	redisClient := redisCli.New(redisCli.SINGLE_MODE, "34.101.216.10:6379", 10,
		redigo.DialReadTimeout(time.Duration(30)*time.Second),
		redigo.DialWriteTimeout(time.Duration(30)*time.Second),
		redigo.DialConnectTimeout(time.Duration(5)*time.Second),
		redigo.DialPassword("skilvulredis"))
	sub = pubsub.NewRedisPubsub(redisClient)

	chString := make(chan string)
	defer close(chString)
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
	userRsc := user.NewAuthCliRsc(auCli, time.Duration(60), time.Duration(30))

	dbInit, err := sqlx.Connect("postgres", "host=34.101.216.10 user=skilvul password=skilvul123apa dbname=skilvul-groupchat sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	dbRsc := chat.NewDBResource(dbInit)
	roomMgr := NewRoomManager(gcClient, dbRsc, userRsc)

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
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
