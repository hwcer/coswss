package coswss

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosnet"
)

var httpServer []*http.Server

func init() {
	cosgo.On(cosgo.EventTypClosing, stop)
}

// stopped 停止coswss模块
// 关闭所有HTTP服务器
func stop() error {
	for _, h := range httpServer {
		_ = h.Close()
	}
	return nil
}

// Handler 绑定各种web框架
func Handler(sockets *cosnet.Sockets, w http.ResponseWriter, r *http.Request) {
	if sockets == nil {
		sockets = cosnet.Default
	}
	h := &handler{sockets: sockets}
	h.ServeHTTP(w, r)
}

// New 启动WebSocket服务器
// address: 监听地址，格式为"host:port"
// route: 路由路径，为空时匹配所有路径
// tlsConfig: TLS配置，用于wss协议
func New(sockets *cosnet.Sockets, address string, route string, tlsConfig ...*tls.Config) (err error) {
	if sockets == nil {
		sockets = cosnet.Default
	}
	h := &handler{route: route, sockets: sockets}
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
			return srv.ListenAndServeTLS("", "") // 使用配置的TLS证书
		} else {
			return srv.ListenAndServe() // 使用HTTP协议
		}
	})
	if errors.Is(err, scc.ErrorTimeout) {
		err = nil // 超时是正常的，因为我们使用了非阻塞的方式启动服务
	}
	return
}

// Accept 使用net.Listener启动WebSocket服务器
// listener: 网络监听器
// route: 路由路径，为空时匹配所有路径
// tlsConfig: TLS配置，用于wss协议
func Accept(sockets *cosnet.Sockets, listener net.Listener, route string, tlsConfig ...*tls.Config) (err error) {
	if sockets == nil {
		sockets = cosnet.Default
	}
	h := &handler{route: route, sockets: sockets}
	srv := &http.Server{
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
			return srv.ServeTLS(listener, "", "") // 使用配置的TLS证书
		} else {
			return srv.Serve(listener) // 使用HTTP协议
		}
	})
	if errors.Is(err, scc.ErrorTimeout) {
		err = nil // 超时是正常的，因为我们使用了非阻塞的方式启动服务
	}
	return
}
