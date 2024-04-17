package coswss

import (
	"bytes"
	"github.com/gorilla/websocket"
	"github.com/hwcer/cosnet/message"
	"log"
	"time"
)

var (
	space   = []byte{' '}
	newline = []byte{'\n'}
)

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
	_, b, err := c.Conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("error: %v", err)
		}
		return nil, err
	}
	b = bytes.TrimSpace(bytes.Replace(b, newline, space, -1))
	msg := message.Require()
	msg.Reset(b)
	return msg, nil
}

func (c *Conn) WriteMessage(msg message.Message) (err error) {
	if c.buff == nil {
		c.buff = new(bytes.Buffer)
	}
	defer func() {
		c.buff.Reset()
	}()

	if _, err = msg.Bytes(c.buff); err != nil {
		return
	}
	return c.Conn.WriteMessage(websocket.BinaryMessage, c.buff.Bytes())
}
