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
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hraban/lush/liblush"
)

var clients liblush.FlexibleMultiWriter

type wrap struct {
	*websocket.Conn
	lock sync.Mutex
}

// Write a message to this websocket client.
func (ws *wrap) Write(data []byte) (int, error) {
	ws.lock.Lock()
	defer ws.lock.Unlock()
	return len(data), ws.WriteMessage(websocket.TextMessage, data)
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	ws := &wrap{Conn: conn}
	clients.AddWriter(ws)
	for {
		typ, msg, err := ws.ReadMessage()
		if err != nil {
			log.Print("Websocket read error: ", err)
			ws.Close()
			clients.RemoveWriter(ws)
			return
		}
		if typ != websocket.TextMessage {
			log.Print("Unexpected websocket message type: ", typ)
			ws.Close()
			clients.RemoveWriter(ws)
			return
		}
		clients.Write(msg)
	}
}

func main() {
	http.HandleFunc("/", handler)
	addr := flag.String("l", "localhost:8081", "listen address")
	flag.Parse()
	log.Print("Starting websocket groupchat server on ", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
