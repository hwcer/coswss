package coswss

import (
	"crypto/tls"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosweb"
	"github.com/hwcer/scc"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func New(so *cosnet.Server) (*Server, error) {
	if so == nil {
		so = cosnet.New(nil)
	}
	ln := &Server{}
	ln.Server = so
	return ln, nil
}

type Server struct {
	*cosnet.Server
	Verify  func(w http.ResponseWriter, r *http.Request) error
	httpSrv *http.Server
}

func (s *Server) connect(w http.ResponseWriter, r *http.Request) error {
	if s.Verify != nil {
		return s.Verify(w, r)
	}
	return nil
}
func (s *Server) HTTPErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(500)
	if r.Method != http.MethodHead {
		_, _ = w.Write([]byte(err.Error()))
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Server.SCC.Stopped() {
		s.HTTPErrorHandler(w, r, errors.New("server is stopped"))
		return
	}
	if err := s.connect(w, r); err != nil {
		s.HTTPErrorHandler(w, r, err)
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err == nil {
		_, err = s.Server.New(&Conn{Conn: conn})
	}
	if err != nil {
		s.HTTPErrorHandler(w, r, err)
	}
}
func (s *Server) handle(c *cosweb.Context, next cosweb.Next) error {
	s.ServeHTTP(c.Response, c.Request)
	return nil
}
func (s *Server) Binding(srv *cosweb.Server, route string) {
	srv.Register(route, s.handle)
}

func (s *Server) Close() (err error) {
	if err = s.Server.Close(); err != nil {
		return
	}
	if s.httpSrv != nil {
		err = s.httpSrv.Close()
	}
	return
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
	return
}
