package coswss

import (
	"bytes"
	"github.com/gorilla/websocket"
	"github.com/hwcer/cosnet/message"
	"github.com/hwcer/logger"
	"io"
	"log"
	"net"
	"time"
)

var (
	space   = []byte{' '}
	newline = []byte{'\n'}
)

func NewConn(c *websocket.Conn) *Conn {
	return &Conn{Conn: c}
}

// Conn net.Conn
type Conn struct {
	*websocket.Conn
	buff *bytes.Buffer
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *Conn) ReadMessage() (message.Message, error) {
	t, b, err := c.Conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err) {
			log.Printf("error: %v", err)
		}
		return nil, err
	}
	if t == websocket.CloseMessage {
		return nil, net.ErrClosed
	}
	if t != websocket.BinaryMessage && t != websocket.TextMessage {
		return nil, nil
	}
	//b = bytes.TrimSpace(bytes.Replace(b, newline, space, -1))
	if len(b) == 0 {
		return nil, io.EOF
	}
	if Transform.Encode != nil {
		if b, err = Transform.Encode(b); err != nil {
			return nil, err
		}
	}

	msg := message.Require()
	if err = msg.Reset(b); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Conn) WriteMessage(msg message.Message) (err error) {
	if c.buff == nil {
		c.buff = new(bytes.Buffer)
	}
	defer func() {
		c.buff.Reset()
	}()

	if _, err = msg.Bytes(c.buff, false); err != nil {
		logger.Error(err)
		return
	}
	b := c.buff.Bytes()
	if len(b) == 0 {
		return
	}
	if Transform.Decode != nil {
		if b, err = Transform.Decode(b); err != nil {
			logger.Error(err)
			return err
		}
	}
	//logger.Trace("Socket response,PATH:%v   BODY:%v", msg.Path(), string(b))
	return c.Conn.WriteMessage(websocket.BinaryMessage, b)
}
