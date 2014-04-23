// Copyright © 2014 Hraban Luyat <hraban@0brg.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hraban/lrucache"
)

// in num msgs not bytes (TODO: should be in bytes)
const BACKLOG_SIZE = 10

// when maximum is reached, room with least recent activity is purged
const NUM_ROOMS = 1000

// https://soundcloud.com/testa-jp/mask-on-mask-re-edit-free-dl
var verbose bool
var files http.Handler
var rooms = lrucache.New(NUM_ROOMS)

// Fuck yeah lrucache
func (cr *chatroom) OnPurge(lrucache.PurgeReason) {
	cr.Close()
}

func handleWebsocket(roomname string, ws *websocket.Conn) {
	obj, err := rooms.Get(roomname)
	if err != nil {
		log.Fatalf("Unexpected error from lrucache.Get(%q): %v", roomname, err)
	}
	cr := obj.(*chatroom)
	c := client(ws)
	id := cr.l_addClient(c)
	for {
		typ, msg, err := c.ReadMessage()
		if err != nil {
			cr.l_delClient(id, c)
			return
		}
		if typ != websocket.TextMessage {
			cr.l_delClient(id, c)
			return
		}
		cr.handleNewMsg(c, msg)
	}
}

func handleNormal(w http.ResponseWriter, r *http.Request) {
	files.ServeHTTP(w, r)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		handleNormal(w, r)
		return
	} else if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", 500)
		return
	} else {
		handleWebsocket(r.URL.Path, ws)
		return
	}
}

func main() {
	rooms.OnMiss(func(roomname string) (lrucache.Cacheable, error) {
		if verbose {
			log.Printf("Created room %q", roomname)
		}
		return newChatroom(BACKLOG_SIZE), nil
	})
	http.HandleFunc("/", handler)
	addr := flag.String("l", "localhost:8081", "listen address")
	flag.BoolVar(&verbose, "v", false, "verbose")
	root := flag.String("root", "", "(optional) root dir for web requests")
	flag.Parse()
	if *root != "" {
		files = http.FileServer(http.Dir(*root))
	} else {
		files = http.NotFoundHandler()
	}
	if verbose {
		log.Print("Starting websocket groupchat server on ", *addr)
	}
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
