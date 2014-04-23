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
	"testing"
)

type clientMock struct{}

func (c clientMock) ReadMessage() (typ int, data []byte, err error) {
	return 0, nil, nil
}
func (c clientMock) WriteMessage(typ int, data []byte) error {
	return nil
}
func (c clientMock) Close() error {
	return nil
}

func TestClientCount(t *testing.T) {
	oldverbose := verbose
	verbose = true
	defer func() {
		verbose = oldverbose
	}()
	cr := newChatroom(10)
	id := cr.addClient(clientMock{})
	if id != 1 {
		t.Errorf("Unexpected id for first client: #%d", id)
	}
	c2 := clientMock{}
	id = cr.addClient(c2)
	if id != 2 {
		t.Errorf("Unexpected id for second client: #%d", id)
	}
	cr.delClient(id, c2)
	if i := cr.numclients.load(); i != 2 {
		t.Error("Unexpected total number of clients:", i)
	}
	if i := cr.connectedclients.load(); i != 1 {
		t.Error("Unexpected number of connected clients:", i)
	}
}
