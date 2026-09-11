package coswss

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hwcer/cosnet"
)

var errTestDenied = errors.New("denied by test")

func newTestHandler(route string) *handler {
	return &handler{route: route, sockets: cosnet.New()}
}

// TestHandlerRouteMismatch 路由不匹配必须 404,不进入升级流程
func TestHandlerRouteMismatch(t *testing.T) {
	h := newTestHandler("/ws")
	r := httptest.NewRequest(http.MethodGet, "http://srv/other", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("路由不匹配应 404,拿到 %d", w.Code)
	}
	if body := w.Body.String(); body == "" {
		t.Fatal("应有 404 响应体")
	}
}

// TestHandlerVerifyRejects Verify 钩子拒绝必须走 HTTPErrorHandler(500)
func TestHandlerVerifyRejects(t *testing.T) {
	old := Options.Verify
	Options.Verify = func(w http.ResponseWriter, r *http.Request) (map[string]string, error) {
		return nil, errTestDenied
	}
	defer func() { Options.Verify = old }()

	h := newTestHandler("")
	r := httptest.NewRequest(http.MethodGet, "http://srv/ws", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Verify 拒绝应 500,拿到 %d", w.Code)
	}
}

// TestHandlerPlainHTTPRejected 带 WS 升级头的请求才走升级;普通 HTTP 请求
// 在 gorilla Upgrade 处被以 400 拒绝(Upgrade 内部已写响应,随后的
// HTTPErrorHandler 500 属多余的二次写,被 net/http 忽略),不得 panic/挂起
func TestHandlerPlainHTTPRejected(t *testing.T) {
	h := newTestHandler("")
	r := httptest.NewRequest(http.MethodGet, "http://srv/ws", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("非 WebSocket 请求应被 400 拒绝,拿到 %d", w.Code)
	}
}

// TestIsWebSocket 升级识别:缺 Upgrade 或缺 Connection: upgrade 都不算
func TestIsWebSocket(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://srv/ws", nil)
	if IsWebSocket(r) {
		t.Fatal("普通请求不应识别为 WebSocket")
	}
	r.Header.Set("Upgrade", "websocket")
	if IsWebSocket(r) {
		t.Fatal("缺 Connection: upgrade 不应识别为 WebSocket")
	}
	r.Header.Set("Connection", "keep-alive, Upgrade")
	if !IsWebSocket(r) {
		t.Fatal("标准升级头应识别为 WebSocket")
	}
}

// TestAccessControlAllow Origin 白名单语义:未配置放行、配置后精确匹配、* 放行
func TestAccessControlAllow(t *testing.T) {
	old := Options.Origin
	defer func() { Options.Origin = old }()

	r := httptest.NewRequest(http.MethodGet, "http://srv/ws", nil)
	r.Header.Set("Origin", "https://game.example.com")

	Options.Origin = nil
	if !AccessControlAllow(r) {
		t.Fatal("未配置白名单应放行")
	}
	Options.Origin = []string{"https://other.example.com"}
	if AccessControlAllow(r) {
		t.Fatal("不在白名单应拒绝")
	}
	Options.Origin = []string{"https://game.example.com"}
	if !AccessControlAllow(r) {
		t.Fatal("精确匹配应放行")
	}
	Options.Origin = []string{"*"}
	if !AccessControlAllow(r) {
		t.Fatal("* 应放行")
	}
}
