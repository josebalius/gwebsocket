package gwebsocket

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrUnexpectedMsgType = errors.New("unexpected message type")
)

type netConn struct {
	wsConn  *websocket.Conn
	msgType int

	reader io.Reader
}

func NewNetConn(wsConn *websocket.Conn, msgType int) *netConn {
	return &netConn{wsConn, msgType, nil}
}

func (c *netConn) Read(b []byte) (int, error) {
	if c.reader == nil {
		msgType, reader, err := c.wsConn.NextReader()
		if err != nil {
			return 0, fmt.Errorf("next reader: %w", err)
		}

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

func (c *netConn) Write(b []byte) (int, error) {
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

func (c *netConn) Close() error {
	return c.wsConn.Close()
}

func (c *netConn) LocalAddr() net.Addr {
	return c.wsConn.LocalAddr()
}

func (c *netConn) RemoteAddr() net.Addr {
	return c.wsConn.RemoteAddr()
}

func (c *netConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}

	return c.SetWriteDeadline(t)
}

func (c *netConn) SetReadDeadline(t time.Time) error {
	return c.wsConn.SetReadDeadline(t)
}

func (c *netConn) SetWriteDeadline(t time.Time) error {
	return c.wsConn.SetWriteDeadline(t)
}
