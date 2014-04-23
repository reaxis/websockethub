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
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/hraban/lrucache"
)

type safeUint32 uint32

// in num msgs not bytes
const BACKLOG_SIZE = 10

// https://soundcloud.com/testa-jp/mask-on-mask-re-edit-free-dl
var backlog = lrucache.New(BACKLOG_SIZE)
var clients struct {
	c map[uint32]*websocket.Conn
	l sync.RWMutex
}
var lowest uint32
var numclients safeUint32
var connectedclients safeUint32
var nummessages safeUint32
var verbose bool
var files http.Handler

func (i *safeUint32) add(x uint32) uint32 {
	return atomic.AddUint32((*uint32)(i), x)
}

func (i *safeUint32) inc() uint32 {
	return i.add(1)
}

func (i *safeUint32) dec() uint32 {
	return i.add(^uint32(0))
}

func (i *safeUint32) load() uint32 {
	return atomic.LoadUint32((*uint32)(i))
}

func extendBacklog(msg []byte) {
	msgid := nummessages.inc()
	if verbose {
	}
	backlog.Set(fmt.Sprint(msgid), msg)
}

func addClient(c *websocket.Conn) uint32 {
	id := numclients.inc()
	clients.l.Lock()
	clients.c[id] = c
	clients.l.Unlock()
	if verbose {
		log.Printf("Client joined: #%d (now: %d)", id, connectedclients.inc())
	}
	return id
}

func delClient(id uint32) {
	clients.l.Lock()
	// Lock is held longer than strictly necessary. Profile before optimizing.
	defer clients.l.Unlock()
	c := clients.c[id]
	c.Close()
	delete(clients.c, id)
	if verbose {
		log.Printf("Client left: #%d (now: %d)", id, connectedclients.dec())
	}
}

func sendToClient(id uint32, c *websocket.Conn, msg []byte) error {
	err := c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		delClient(id)
		return err
	}
	return nil
}

func handleWebsocket(ws *websocket.Conn) {
	var i uint32
	if nummessages.load() < BACKLOG_SIZE {
		i = 0
	} else {
		i = nummessages.load() - BACKLOG_SIZE
	}
	for i <= nummessages.load() {
		msgi, err := backlog.Get(fmt.Sprint(i))
		if err != lrucache.ErrNotFound {
			msg := msgi.([]byte)
			if ws.WriteMessage(websocket.TextMessage, msg) != nil {
				return
			}
		}
		i += 1
	}
	id := addClient(ws)
	for {
		typ, msg, err := ws.ReadMessage()
		if err != nil {
			delClient(id)
			return
		}
		if typ != websocket.TextMessage {
			delClient(id)
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
		http.Error(w, "internal server error", 500)
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
