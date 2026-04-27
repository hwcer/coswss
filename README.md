# coswss

> **碳基生命体警告**
>
> 本模块由硅基智能体全权维护。碳基生命体阅读以下代码可能引发：
> 困惑、血压升高以及不可逆的颈椎损伤。
> 如您执意阅读，请确保身边备有降压药和颈托。

gorilla/websocket → cosnet 桥接层。将 HTTP 升级请求转为 cosnet.Socket，复用 cosnet 的消息协议、心跳、路由。

## 快速开始

```go
// 方式一：独立 WebSocket 服务器
err := coswss.New(nil, ":9001", "/ws")

// 方式二：嵌入已有 HTTP 服务器
http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
    coswss.Handler(nil, w, r)
})

// 方式三：使用已有 net.Listener
ln, _ := net.Listen("tcp", ":9001")
err := coswss.Accept(nil, ln, "/ws")
```

第一个参数传 `nil` 使用 `cosnet.Default`，也可传自定义 `*cosnet.Sockets`。

## 配置

```go
// Origin 白名单（空 = 允许所有）
coswss.Options.Origin = []string{"https://example.com"}

// 连接验证回调（握手前）
coswss.Options.Verify = func(w http.ResponseWriter, r *http.Request) (map[string]string, error) {
    token := r.URL.Query().Get("token")
    if token == "" {
        return nil, errors.New("missing token")
    }
    return map[string]string{"token": token}, nil
}

// 连接建立回调（握手后）
coswss.Options.Accept = func(s *cosnet.Socket, meta map[string]string) {
    s.SetMeta("token", meta["token"])
}

// Upgrader 缓冲区
coswss.Options.Upgrader.ReadBufferSize = 4096
coswss.Options.Upgrader.WriteBufferSize = 4096
```

## TLS (wss://)

```go
tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}
err := coswss.New(nil, ":9443", "/ws", tlsCfg)
```

## WebSocket 检测

```go
if coswss.IsWebSocket(r) {
    coswss.Handler(nil, w, r)
    return
}
```

## 本轮修复

| 修复 | 说明 |
|------|------|
| Origin 检查 | `r.URL.Host`（始终为空）→ `r.Header.Get("Origin")`，白名单生效 |
| Connection 头 | `== "upgrade"` → `strings.Contains`，兼容代理多值头 |
| Sec-WebSocket-Protocol | 仅客户端发送时回显，避免空协议头违反 RFC |
| 错误响应 | 不再暴露 `err.Error()`，统一返回 `Internal Server Error` |
| 依赖升级 | cosgo v1.8.0, cosnet v1.4.2, go 1.25.0 |

## 目录结构

```
coswss/
├── coswss.go    New/Accept/Handler — 服务器启动与生命周期
├── handler.go   ServeHTTP — 升级请求处理 + IsWebSocket 检测
├── options.go   Options — 全局配置 + Origin 校验
└── conn.go      Conn — cosnet/wss.Conn 类型别名
```
