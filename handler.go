package coswss

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/logger"
)

// handler 处理WebSocket请求

type handler struct {
	route   string // 路由路径
	sockets *cosnet.Sockets
}

func (s *handler) HTTPErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	logger.Alert(err)
	w.WriteHeader(http.StatusInternalServerError)
	if r.Method != http.MethodHead {
		_, _ = w.Write([]byte("Internal Server Error"))
	}
}

// ServeHTTP 处理WebSocket连接请求
func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if scc.Stopped() {
		s.HTTPErrorHandler(w, r, errors.New("server is stopped"))
		return
	}
	if s.route != "" && r.URL.Path != s.route {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 page not found"))
		return
	}

	var err error
	var meta map[string]string
	if Options.Verify != nil {
		meta, err = Options.Verify(w, r)
	}
	if err != nil {
		s.HTTPErrorHandler(w, r, err)
		return
	}

	var header http.Header
	// 🔴 RFC 6455 §4.2.2:服务端必须从客户端子协议列表中选定**一个**回显。
	// 旧实现原样回显整个逗号列表,gorilla(Upgrader.Subprotocols 为空时)照抄
	// responseHeader——请求 "chat, superchat" 的客户端收到的是逗号列表而非单 token,
	// 浏览器校验失败即握手失败。使用方配置了 Upgrader.Subprotocols 时 gorilla
	// 自带交集选择并忽略 responseHeader,此处不再干预。
	// 🔴 "已配置"判定必须与 gorilla 同口径(Subprotocols != nil):用户显式赋
	// []string{}(非 nil 空)时,按 len==0 判定会设置回显头,而 gorilla 走交集路径
	// 空交集回空串把该头丢弃——两头语义错位,握手退化为无子协议
	if Options.Upgrader.Subprotocols == nil {
		if subs := websocket.Subprotocols(r); len(subs) > 0 {
			header = http.Header{"Sec-WebSocket-Protocol": {subs[0]}}
		}
	}

	conn, err := Options.Upgrader.Upgrade(w, r, header)
	if err != nil {
		s.HTTPErrorHandler(w, r, err)
		return
	}
	var sock *cosnet.Socket
	sock, err = s.sockets.Create(NewConn(conn))
	if err != nil {
		//Upgrade成功后连接已hijack,再写HTTP状态码无效,直接关闭连接
		logger.Alert(err)
		_ = conn.Close()
		return
	}
	if Options.Accept != nil {
		Options.Accept(sock, meta)
	}
}

func IsWebSocket(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}
