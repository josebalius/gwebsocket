package gwebsocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestNewConn(t *testing.T) {
	wsConn := websocket.Conn{}
	nc := NewNetConn(&wsConn, websocket.TextMessage)
	if nc == nil {
		t.Error("netConn is nil")
	}
}

func TestConn(t *testing.T) {
	upgrader := websocket.Upgrader{}
	ping := []byte("ping")
	pong := []byte("pong")

	done := make(chan error)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		c, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			done <- fmt.Errorf("upgrade req: %w", err)
			return
		}
		defer c.Close()

		nc := NewNetConn(c, websocket.TextMessage)
		b := make([]byte, len(ping))
		n, err := nc.Read(b)
		if err != nil {
			done <- fmt.Errorf("server read message: %w", err)
			return
		}
		if n != len(ping) {
			t.Errorf("reading ping message len unexpected, got '%v'", n)
		}

		msg := string(b)
		if msg != "ping" {
			done <- fmt.Errorf("incoming message is not expected value, got: '%v'", msg)
			return
		}

		n, err = nc.Write(pong)
		if err != nil {
			done <- fmt.Errorf("write pong: %w", err)
			return
		}
		if n != len(pong) {
			done <- fmt.Errorf("pong message len unexpected, got: '%v'", n)
			return
		}

		done <- nil
	}))
	defer server.Close()

	c, _, err := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http", "ws", -1), nil)
	if err != nil {
		t.Errorf("dial: %w", err)
	}
	defer c.Close()

	nc := NewNetConn(c, websocket.TextMessage)

	n, err := nc.Write(ping)
	if err != nil {
		t.Errorf("write message: %w", err)
	}
	if n != len(ping) {
		t.Errorf("ping message len unexpected, got: '%v'", n)
	}

	b := make([]byte, len(pong))
	n, err = nc.Read(b)
	if err != nil {
		t.Errorf("read message: %w", err)
	}
	if n != len(pong) {
		t.Errorf("reading pong message len unexpected, got '%v'", n)
	}

	msg := string(b)
	if msg != "pong" {
		t.Errorf("client incoming message unexpectd, got: '%v'", msg)
	}

	if err := <-done; err != nil {
		t.Error(err)
	}
}
