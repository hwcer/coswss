package coswss

import (
	"crypto/tls"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosweb"
	"net/http"
	"sync/atomic"
	"time"
)

var started int32
var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var httpServer []*http.Server

func init() {
	upgrader.CheckOrigin = AccessControlAllow
}

func AccessControlAllow(r *http.Request) bool {
	if len(Options.Origin) == 0 {
		return true
	}
	for _, o := range Options.Origin {
		if o == "*" || o == r.URL.Host {
			return true
		}
	}
	return false
}

func start() {
	if atomic.CompareAndSwapInt32(&started, 0, 1) {
		scc.Trigger(stopped)
	}

}
func stopped() {
	for _, h := range httpServer {
		_ = h.Close()
	}
}

func Binding(srv *cosweb.Server, route string) error {
	h := &handler{}
	srv.Register(route, h.handle)
	return nil
}

func Listen(address string, route string, tlsConfig ...*tls.Config) (err error) {
	h := &handler{route: route}
	srv := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           h,
	}
	httpServer = append(httpServer, srv)
	if len(tlsConfig) > 0 {
		srv.TLSConfig = tlsConfig[0]
	}
	//启动服务
	err = scc.Timeout(time.Second, func() error {
		if srv.TLSConfig != nil {
			return srv.ListenAndServeTLS("", "")
		} else {
			return srv.ListenAndServe()
		}
	})
	if errors.Is(err, scc.ErrorTimeout) {
		err = nil
	}
	if err == nil {
		start()
	}
	return
}
