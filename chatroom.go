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
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hraban/lrucache"
)

type client interface {
	ReadMessage() (typ int, data []byte, err error)
	WriteMessage(typ int, data []byte) error
	Close() error
}

type chatroom struct {
	backlog     *lrucache.Cache
	nummessages safeUint32
	// Only used in verbose mode
	numclients       safeUint32
	connectedclients safeUint32
	// unsafe for concurrent use
	l struct {
		clientsL sync.RWMutex
		clients  map[uint32]client
	}
}

func newChatroom(backlogsize int64) *chatroom {
	cr := &chatroom{}
	cr.backlog = lrucache.New(BACKLOG_SIZE)
	cr.l.clients = map[uint32]client{}
	return cr
}

func (cr *chatroom) extendBacklog(msg []byte) {
	msgid := cr.nummessages.inc()
	cr.backlog.Set(fmt.Sprint(msgid), msg)
}

func (cr *chatroom) sendBacklog(c client) {
	var i uint32
	if cr.nummessages.load() < BACKLOG_SIZE {
		i = 0
	} else {
		i = cr.nummessages.load() - BACKLOG_SIZE
	}
	for i <= cr.nummessages.load() {
		msgi, err := cr.backlog.Get(fmt.Sprint(i))
		if err != lrucache.ErrNotFound {
			msg := msgi.([]byte)
			if c.WriteMessage(websocket.TextMessage, msg) != nil {
				return
			}
		}
		i += 1
	}
}

func (cr *chatroom) addClient(c client) uint32 {
	id := cr.numclients.inc()
	cr.l.clientsL.Lock()
	cr.l.clients[id] = c
	cr.l.clientsL.Unlock()
	cr.sendBacklog(c)
	if verbose {
		total := cr.connectedclients.inc()
		log.Printf("Client joined: #%d (now: %d)", id, total)
	}
	return id
}

func (cr *chatroom) delClient(id uint32, c client) {
	cr.l.clientsL.Lock()
	// Lock is held longer than strictly necessary. Profile before optimizing.
	defer cr.l.clientsL.Unlock()
	delete(cr.l.clients, id)
	c.Close()
	if verbose {
		total := cr.connectedclients.dec()
		log.Printf("Client left: #%d (now: %d)", id, total)
	}
}

func (cr *chatroom) sendToClient(id uint32, c client, msg []byte) error {
	err := c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		cr.delClient(id, c)
		return err
	}
	return nil
}

// send message to each client, in a separate goroutine
func (cr *chatroom) sendToAllClientsAsync(msg []byte) {
	cr.l.clientsL.RLock()
	for id, c := range cr.l.clients {
		go cr.sendToClient(id, c, msg)
	}
	cr.l.clientsL.RUnlock()
}

func (cr *chatroom) handleNewMsg(from client, msg []byte) {
	go cr.extendBacklog(msg)
	cr.sendToAllClientsAsync(msg)
}
