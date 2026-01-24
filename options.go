package coswss

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hwcer/cosnet"
)

// Options coswss模块配置选项
var Options = struct {
	// Accept 连接建立后的回调函数
	Accept func(s *cosnet.Socket, meta map[string]string)
	// Verify 连接建立前的验证函数
	Verify func(w http.ResponseWriter, r *http.Request) (meta map[string]string, err error)
	// Origin 允许的来源域名列表
	Origin []string
	// Upgrader websocket升级器配置
	Upgrader websocket.Upgrader
}{
	Upgrader: websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}, // 默认配置
}

func init() {
	Options.Upgrader.CheckOrigin = AccessControlAllow
}

// AccessControlAllow 用于检查WebSocket连接的来源是否合法
func AccessControlAllow(r *http.Request) bool {
	if len(Options.Origin) == 0 {
		return true // 默认允许所有来源，实际生产环境中应该根据需要进行限制
	}
	for _, o := range Options.Origin {
		if o == "*" || o == r.URL.Host {
			return true
		}
	}
	return false
}
