package gwebsocket

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

var _ net.Conn = (*NetConn)(nil)

var (
	// ErrUnexpectedMsgType is returned when a reader is returned by the websocket
	// connection that does not match the message type NetConn was created with
	ErrUnexpectedMsgType = errors.New("unexpected message type")
)

type NetConn struct {
	wsConn  *websocket.Conn
	msgType int

	reader io.Reader
}

// NewNetConn returns a NetConn pointer. It takes a gorilla websocket
// connection and a message type. See: https://github.com/gorilla/websocket/blob/v1.4.2/conn.go#L62
// for possible values
func NewNetConn(wsConn *websocket.Conn, msgType int) *NetConn {
	return &NetConn{wsConn, msgType, nil}
}

func (c *NetConn) Read(b []byte) (int, error) {
	if c.reader == nil {
		msgType, reader, err := c.wsConn.NextReader()
		if err != nil {
			return 0, fmt.Errorf("next reader: %w", err)
		}

		// err if we receive an unsupported message type
		if msgType != c.msgType {
			return 0, ErrUnexpectedMsgType
		}

		c.reader = reader
	}

	bytesRead, err := c.reader.Read(b)
	if err != nil {
		c.reader = nil

		if err == io.EOF {
			err = nil
		}
	}

	return bytesRead, err
}

func (c *NetConn) Write(b []byte) (int, error) {
	nextWriter, err := c.wsConn.NextWriter(c.msgType)
	if err != nil {
		return 0, fmt.Errorf("next writer: %w", err)
	}

	bytesWritten, err := nextWriter.Write(b)
	if err != nil {
		return bytesWritten, fmt.Errorf("write: %w", err)
	}

	return bytesWritten, nextWriter.Close()
}

func (c *NetConn) Close() error {
	return c.wsConn.Close()
}

func (c *NetConn) LocalAddr() net.Addr {
	return c.wsConn.LocalAddr()
}

func (c *NetConn) RemoteAddr() net.Addr {
	return c.wsConn.RemoteAddr()
}

func (c *NetConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}

	return c.SetWriteDeadline(t)
}

func (c *NetConn) SetReadDeadline(t time.Time) error {
	return c.wsConn.SetReadDeadline(t)
}

func (c *NetConn) SetWriteDeadline(t time.Time) error {
	return c.wsConn.SetWriteDeadline(t)
}
