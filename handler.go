package coswss

import (
	"errors"
	"net/http"

	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/logger"
)

// handler 处理WebSocket请求

type handler struct {
	route string // 路由路径
}

// HTTPErrorHandler 处理HTTP错误
func (s *handler) HTTPErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(500)
	if r.Method != http.MethodHead {
		_, _ = w.Write([]byte(err.Error()))
	}
	logger.Alert(err)
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

	var header = map[string][]string{"Sec-WebSocket-Protocol": {r.Header.Get("Sec-WebSocket-Protocol")}}

	conn, err := Options.Upgrader.Upgrade(w, r, header)
	if err != nil {
		s.HTTPErrorHandler(w, r, err)
		return
	}
	var sock *cosnet.Socket
	sock, err = cosnet.New(NewConn(conn))
	if err != nil {
		s.HTTPErrorHandler(w, r, err)
		return
	}
	if Options.Accept != nil {
		Options.Accept(sock, meta)
	}
}
