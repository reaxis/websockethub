// Copyright © 2013 Hraban Luyat <hraban@0brg.net>
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
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/hraban/lrucache"
)

// in num msgs not bytes
const BACKLOG_SIZE = 10

// https://soundcloud.com/testa-jp/mask-on-mask-re-edit-free-dl
var backlog = lrucache.New(BACKLOG_SIZE)
var clients struct {
	c map[uint32]*websocket.Conn
	l sync.RWMutex
}
var lowest uint32
var numclients uint32
var connectedclients int32
var nummessages uint32
var verbose bool
var files http.Handler

func inc(i *uint32) uint32 {
	return atomic.AddUint32(i, 1)
}

func extendBacklog(msg []byte) {
	msgid := inc(&nummessages)
	if verbose {
	}
	backlog.Set(fmt.Sprint(msgid), msg)
}

func sendToClient(id uint32, c *websocket.Conn, msg []byte) error {
	err := c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		clients.l.Lock()
		delete(clients.c, id)
		clients.l.Unlock()
		if verbose {
			atomic.AddInt32(&connectedclients, -1)
			log.Printf("Client count -1: %d\n", connectedclients)
		}
		return err
	}
	return nil
}

func ld(i *uint32) uint32 {
	return atomic.LoadUint32(i)
}

func handleWebsocket(ws *websocket.Conn) {
	if verbose {
		atomic.AddInt32(&connectedclients, 1)
		log.Printf("Client count +1: %d\n", connectedclients)
	}
	var i uint32
	if ld(&nummessages) < BACKLOG_SIZE {
		i = 0
	} else {
		i = ld(&nummessages) - BACKLOG_SIZE
	}
	for i <= ld(&nummessages) {
		msgi, err := backlog.Get(fmt.Sprint(i))
		if err != lrucache.ErrNotFound {
			msg := msgi.([]byte)
			if ws.WriteMessage(websocket.TextMessage, msg) != nil {
				return
			}
		}
		inc(&i)
	}
	id := inc(&numclients)
	clients.l.Lock()
	clients.c[id] = ws
	clients.l.Unlock()
	for {
		typ, msg, err := ws.ReadMessage()
		if err != nil {
			ws.Close()
			return
		}
		if typ != websocket.TextMessage {
			ws.Close()
			return
		}
		go extendBacklog(msg)
		clients.l.RLock()
		for id, c := range clients.c {
			go sendToClient(id, c, msg)
		}
		clients.l.RUnlock()
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
		return
	} else {
		handleWebsocket(ws)
		return
	}
}

func main() {
	clients.c = map[uint32]*websocket.Conn{}
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
