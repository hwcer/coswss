# coswss 模块

coswss 是一个基于 Golang 的 WebSocket 服务器实现模块，提供了简单易用的 WebSocket 服务器启动和管理功能。

## 功能特点

- **简单易用**：提供了简洁的 API 接口，方便快速启动 WebSocket 服务器
- **多种启动方式**：支持通过 HTTP 服务器启动，也支持通过 net.Listener 启动
- **配置灵活**：支持自定义 Upgrader 配置，如 ReadBufferSize、WriteBufferSize 等
- **安全性**：支持设置 CheckOrigin 函数，控制跨域请求
- **错误处理**：提供了 HTTPErrorHandler 接口，方便自定义错误处理逻辑

## 模块结构

- **coswss.go**：核心模块，包含 WebSocket 服务器的启动和管理功能
- **handler.go**：处理器模块，处理 WebSocket 连接请求
- **options.go**：配置选项模块，定义了 coswss 模块的配置选项
- **conn.go**：连接模块，封装了 WebSocket 连接

## 使用示例

### 1. 基本用法

```go
import (
    "github.com/hwcer/coswss"
    "net/http"
)

func main() {
    // 启动 WebSocket 服务器
    if err := coswss.Listen(":8080", "/ws"); err != nil {
        panic(err)
    }
    
    // 启动 HTTP 服务器
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })
    
    http.ListenAndServe(":8080", nil)
}
```

### 2. 使用 net.Listener 启动

```go
import (
    "github.com/hwcer/coswss"
    "net"
)

func main() {
    // 创建 TCP 监听器
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        panic(err)
    }
    defer listener.Close()
    
    // 使用 net.Listener 启动 WebSocket 服务器
    if err := coswss.NewWithListener(listener, "/ws"); err != nil {
        panic(err)
    }
    
    // 等待连接
    select {}
}
```

### 3. 自定义配置

```go
import (
    "github.com/hwcer/coswss"
    "net/http"
)

func main() {
    // 自定义 Upgrader 配置
    coswss.Options.Upgrader.ReadBufferSize = 1024
    coswss.Options.Upgrader.WriteBufferSize = 1024
    
    // 自定义 CheckOrigin 函数
    coswss.Options.Upgrader.CheckOrigin = func(r *http.Request) bool {
        return true // 允许所有跨域请求
    }
    
    // 启动 WebSocket 服务器
    if err := coswss.Listen(":8080", "/ws"); err != nil {
        panic(err)
    }
    
    // 启动 HTTP 服务器
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })
    
    http.ListenAndServe(":8080", nil)
}
```

### 4. 自定义错误处理

```go
import (
    "github.com/hwcer/coswss"
    "net/http"
)

func main() {
    // 自定义 HTTP 错误处理函数
    coswss.Options.HTTPErrorHandler = func(w http.ResponseWriter, r *http.Request, status int) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        w.Write([]byte(`{"error": "WebSocket connection failed"}`))
    }
    
    // 启动 WebSocket 服务器
    if err := coswss.Listen(":8080", "/ws"); err != nil {
        panic(err)
    }
    
    // 启动 HTTP 服务器
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })
    
    http.ListenAndServe(":8080", nil)
}
```

## 配置选项

coswss 模块的配置选项定义在 `Options` 变量中，包含以下字段：

- **Upgrader**：WebSocket 连接升级器，包含 ReadBufferSize、WriteBufferSize、CheckOrigin 等配置
- **HTTPErrorHandler**：HTTP 错误处理函数，用于处理 WebSocket 连接失败时的错误

## API 接口

### 1. Listen

```go
func Listen(addr, route string, tlsConfig ...*tls.Config) (err error)
```

- **addr**：服务器地址，如 ":8080"
- **route**：WebSocket 路由路径，如 "/ws"
- **tlsConfig**：可选的 TLS 配置，用于启用 HTTPS
- **返回值**：错误信息

### 2. NewWithListener

```go
func NewWithListener(listener net.Listener, route string, tlsConfig ...*tls.Config) (err error)
```

- **listener**：网络监听器，如 TCP 监听器
- **route**：WebSocket 路由路径，如 "/ws"
- **tlsConfig**：可选的 TLS 配置，用于启用 HTTPS
- **返回值**：错误信息

### 3. Stop

```go
func Stop()
```

- 停止 WebSocket 服务器

### 4. Started

```go
func Started() bool
```

- **返回值**：WebSocket 服务器是否已启动

## 注意事项

- **路由路径**：WebSocket 路由路径应该以 "/" 开头，如 "/ws"
- **端口冲突**：如果 WebSocket 服务器和 HTTP 服务器使用同一个端口，确保它们的路由路径不冲突
- **安全性**：在生产环境中，应该设置合适的 CheckOrigin 函数，避免跨域攻击
- **性能**：根据实际需求，调整 Upgrader 的 ReadBufferSize 和 WriteBufferSize 配置，以优化性能
