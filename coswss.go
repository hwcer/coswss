package coswss

import (
	"crypto/tls"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosweb"
	"net/http"
	"sync/atomic"
	"time"
)

func New() *Server {
	ln := &Server{}
	ln.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ln.upgrader.CheckOrigin = ln.AccessControlAllow
	return ln
}

type Server struct {
	Accept   func(s *cosnet.Socket, uid string)
	Verify   func(w http.ResponseWriter, r *http.Request) (uid string, err error)
	Origin   []string
	httpSrv  *http.Server
	started  int32
	upgrader websocket.Upgrader
}

func (s *Server) HTTPErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(500)
	if r.Method != http.MethodHead {
		_, _ = w.Write([]byte(err.Error()))
	}
	logger.Alert(err)
}

func (s *Server) AccessControlAllow(r *http.Request) bool {
	if len(s.Origin) == 0 {
		return true
	}
	for _, o := range s.Origin {
		if o == "*" || o == r.URL.Host {
			return true
		}
	}
	return false
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if scc.Stopped() {
		s.HTTPErrorHandler(w, r, errors.New("server is stopped"))
		return
	}
	var err error
	var uid string
	if s.Verify != nil {
		uid, err = s.Verify(w, r)
	}
	if err != nil {
		s.HTTPErrorHandler(w, r, err)
		return
	}

	var header = map[string][]string{"Sec-WebSocket-Protocol": {r.Header.Get("Sec-WebSocket-Protocol")}}

	conn, err := s.upgrader.Upgrade(w, r, header)
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
	if s.Accept != nil {
		s.Accept(sock, uid)
	}
}
func (s *Server) handle(c *cosweb.Context, next cosweb.Next) error {
	s.ServeHTTP(c.Response, c.Request)
	return nil
}
func (s *Server) Binding(srv *cosweb.Server, route string) {
	srv.Register(route, s.handle)
	s.start()
}

func (s *Server) start() {
	if atomic.CompareAndSwapInt32(&s.started, 0, 1) {
		scc.Trigger(s.stopped)
	}

}
func (s *Server) stopped() {
	if s.httpSrv != nil {
		_ = s.httpSrv.Close()
	}
}

func (s *Server) Start(address string, tlsConfig ...*tls.Config) (err error) {
	srv := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           s,
	}
	s.httpSrv = srv
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
		s.start()
	}
	return
}
