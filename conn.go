package coswss

import (
	"github.com/gorilla/websocket"
	"github.com/hwcer/cosnet/wss"
)

// NewConn 创建一个新的WebSocket连接
func NewConn(c *websocket.Conn) *wss.Conn {
	return wss.NewConn(c)
}

// Conn 是cosnet/wss.Conn的别名
// 使用cosnet/wss中的Conn实现，避免代码重复
type Conn = wss.Conn
