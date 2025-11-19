# fursy plugins/stream

Stream integration plugin for fursy router. Provides seamless SSE and WebSocket support using [github.com/coregx/stream](https://github.com/coregx/stream) library.

## Features

- **SSE Hub Middleware**: Share SSE hub across handlers with type-safe generics
- **WebSocket Hub Middleware**: Share WebSocket hub across handlers
- **Context Helpers**: `stream.SSEUpgrade()` and `stream.WebSocketUpgrade()` for easy connection upgrades
- **Type-safe Hub Retrieval**: Generic helpers `GetSSEHub[T]()` and `GetWebSocketHub()` for hub access
- **Production Ready**: Built on battle-tested [stream v0.1.0](https://github.com/coregx/stream) (314 tests, 84.3% coverage)

## Installation

```bash
go get github.com/coregx/fursy/plugins/stream
go get github.com/coregx/stream
```

## Quick Start

### Server-Sent Events (SSE)

```go
package main

import (
    "log"
    "time"

    "github.com/coregx/fursy"
    "github.com/coregx/fursy/plugins/stream"
    "github.com/coregx/stream/sse"
)

type Notification struct {
    Type    string    `json:"type"`
    Message string    `json:"message"`
    Time    time.Time `json:"time"`
}

func main() {
    // Create SSE Hub.
    hub := sse.NewHub[Notification]()
    go hub.Run()
    defer hub.Close()

    // Create router.
    router := fursy.New()
    router.Use(stream.SSEHub(hub))

    // SSE endpoint.
    router.GET("/events", func(c *fursy.Context) error {
        hub, _ := stream.GetSSEHub[Notification](c)

        return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
            hub.Register(conn)
            defer hub.Unregister(conn)
            <-conn.Done()
            return nil
        })
    })

    // Broadcast endpoint.
    router.POST("/notify", func(c *fursy.Context) error {
        hub, _ := stream.GetSSEHub[Notification](c)

        var notification Notification
        if err := c.Bind(&notification); err != nil {
            return c.Problem(fursy.BadRequest(err.Error()))
        }

        notification.Time = time.Now()
        hub.BroadcastJSON(notification)

        return c.JSON(200, map[string]string{"status": "sent"})
    })

    log.Fatal(router.Run(":8080"))
}
```

**Client usage:**
```bash
# Listen for events
curl -N http://localhost:8080/events

# Send notification (in another terminal)
curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{"type":"info","message":"Hello, SSE!"}'
```

### WebSocket

```go
package main

import (
    "log"

    "github.com/coregx/fursy"
    "github.com/coregx/fursy/plugins/stream"
    "github.com/coregx/stream/websocket"
)

func main() {
    // Create WebSocket Hub.
    hub := websocket.NewHub()
    go hub.Run()
    defer hub.Close()

    // Create router.
    router := fursy.New()
    router.Use(stream.WebSocketHub(hub))

    // WebSocket endpoint.
    router.GET("/ws", func(c *fursy.Context) error {
        hub, _ := stream.GetWebSocketHub(c)

        return stream.WebSocketUpgrade(c, func(conn *websocket.Conn) error {
            hub.Register(conn)
            defer hub.Unregister(conn)

            // Read loop.
            for {
                msgType, data, err := conn.Read()
                if err != nil {
                    return err
                }

                // Broadcast to all clients.
                hub.Broadcast(data)
            }
        }, nil)
    })

    // Health check.
    router.GET("/health", func(c *fursy.Context) error {
        hub, _ := stream.GetWebSocketHub(c)
        return c.JSON(200, map[string]any{
            "status":  "ok",
            "clients": hub.ClientCount(),
        })
    })

    log.Fatal(router.Run(":8080"))
}
```

**Client usage:**
```bash
# Connect with wscat
wscat -c ws://localhost:8080/ws

# Type messages - they'll be broadcast to all connected clients
```

## API Reference

### SSE Middleware

#### `SSEHub[T any](hub *sse.Hub[T]) fursy.HandlerFunc`

Creates a middleware that provides SSE Hub in request context.

Type parameter `T` specifies the type of events that will be broadcast through the hub.

**Example:**
```go
type MyEvent struct {
    ID      int    `json:"id"`
    Message string `json:"message"`
}

hub := sse.NewHub[MyEvent]()
go hub.Run()
defer hub.Close()

router.Use(stream.SSEHub(hub))
```

#### `GetSSEHub[T any](c *fursy.Context) (*sse.Hub[T], bool)`

Retrieves SSE hub from request context.

Returns `(hub, true)` if hub is found and has the correct type T.
Returns `(nil, false)` if hub not found or type mismatch.

**Example:**
```go
hub, ok := stream.GetSSEHub[MyEvent](c)
if !ok {
    return c.Problem(fursy.InternalServerError("Hub not configured"))
}

hub.BroadcastJSON(MyEvent{ID: 1, Message: "Hello"})
```

#### `SSEUpgrade(c *fursy.Context, handler func(conn *sse.Conn) error) error`

Upgrades HTTP connection to Server-Sent Events.

The handler function receives an SSE connection and should handle the SSE lifecycle.
The connection is automatically closed when the handler returns.

**Example:**
```go
return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
    hub.Register(conn)
    defer hub.Unregister(conn)
    <-conn.Done()
    return nil
})
```

### WebSocket Middleware

#### `WebSocketHub(hub *websocket.Hub) fursy.HandlerFunc`

Creates a middleware that provides WebSocket Hub in request context.

**Example:**
```go
hub := websocket.NewHub()
go hub.Run()
defer hub.Close()

router.Use(stream.WebSocketHub(hub))
```

#### `GetWebSocketHub(c *fursy.Context) (*websocket.Hub, bool)`

Retrieves WebSocket hub from request context.

Returns `(hub, true)` if hub is found.
Returns `(nil, false)` if hub not found.

**Example:**
```go
hub, ok := stream.GetWebSocketHub(c)
if !ok {
    return c.Problem(fursy.InternalServerError("Hub not configured"))
}

hub.Broadcast([]byte("Hello, WebSocket!"))
```

#### `WebSocketUpgrade(c *fursy.Context, handler func(conn *websocket.Conn) error, opts *websocket.UpgradeOptions) error`

Upgrades HTTP connection to WebSocket.

The handler function receives a WebSocket connection and should handle the WebSocket lifecycle.
The connection is automatically closed when the handler returns.

**Example:**
```go
return stream.WebSocketUpgrade(c, func(conn *websocket.Conn) error {
    hub.Register(conn)
    defer hub.Unregister(conn)

    for {
        msgType, data, err := conn.Read()
        if err != nil {
            return err
        }
        hub.Broadcast(data)
    }
}, nil)
```

## Advanced Usage

### Custom Upgrade Options (WebSocket)

```go
opts := &websocket.UpgradeOptions{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        // Allow connections from specific origins only.
        origin := r.Header.Get("Origin")
        return origin == "https://example.com"
    },
}

return stream.WebSocketUpgrade(c, handler, opts)
```

### SSE with Custom Event Types

```go
return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
    hub.Register(conn)
    defer hub.Unregister(conn)

    // Send custom event.
    event := sse.NewEvent().
        WithType("custom-event").
        WithData("custom data").
        WithID("event-123").
        WithRetry(5000)

    if err := conn.Send(event); err != nil {
        return err
    }

    <-conn.Done()
    return nil
})
```

### Broadcasting to Specific Clients (WebSocket)

```go
// Broadcast to all except sender.
hub.BroadcastExcept(data, senderConn)

// Send to specific client only.
if err := targetConn.Write(websocket.TextMessage, data); err != nil {
    // Handle error.
}
```

## Examples

Full working examples are available in `examples/` directory:

- **sse-notifications**: SSE notification server with POST broadcast endpoint
- **websocket-chat**: WebSocket chat server with multiple clients

## Performance

Built on top of [stream v0.1.0](https://github.com/coregx/stream) which delivers:

- **SSE Hub**: 4.7 μs/op for 10 clients, 43.6 μs/op for 100 clients
- **WebSocket Hub**: 11 μs/op for 10 clients, 75 μs/op for 100 clients
- **Zero external dependencies**: Pure stdlib implementation
- **Production tested**: 314 tests, 84.3% coverage

## Documentation

- [SSE Guide](https://github.com/coregx/stream/blob/main/docs/SSE_GUIDE.md) - Complete SSE documentation
- [WebSocket Guide](https://github.com/coregx/stream/blob/main/docs/WEBSOCKET_GUIDE.md) - Complete WebSocket documentation
- [stream Library](https://github.com/coregx/stream) - Source library documentation

## License

MIT License - see [LICENSE](../../LICENSE) file for details.

## Contributing

Contributions are welcome! Please see [fursy CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## Support

- **Issues**: [github.com/coregx/fursy/issues](https://github.com/coregx/fursy/issues)
- **Discussions**: [github.com/coregx/fursy/discussions](https://github.com/coregx/fursy/discussions)

## See Also

- [fursy Router](../../README.md) - Main router documentation
- [plugins/database](../database/README.md) - Database integration
- [stream library](https://github.com/coregx/stream) - Standalone SSE + WebSocket
- [Examples](../../examples/)
  - [SSE Notifications](../../examples/07-sse-notifications/)
  - [WebSocket Chat](../../examples/08-websocket-chat/)

## Related Guides

- [SSE Guide](https://github.com/coregx/stream/blob/main/docs/SSE_GUIDE.md) - Complete SSE documentation
- [WebSocket Guide](https://github.com/coregx/stream/blob/main/docs/WEBSOCKET_GUIDE.md) - Complete WebSocket documentation
