package coswss

import (
	"errors"
	"net/http"
	"strings"

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
	if proto := r.Header.Get("Sec-WebSocket-Protocol"); proto != "" {
		header = http.Header{"Sec-WebSocket-Protocol": {proto}}
	}

	conn, err := Options.Upgrader.Upgrade(w, r, header)
	if err != nil {
		s.HTTPErrorHandler(w, r, err)
		return
	}
	var sock *cosnet.Socket
	sock, err = s.sockets.Create(NewConn(conn))
	if err != nil {
		s.HTTPErrorHandler(w, r, err)
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
