# fursy Examples

This directory contains comprehensive examples demonstrating fursy features.

## Quick Navigation

| Example | Description | Plugins | Complexity |
|---------|-------------|---------|------------|
| [01-hello-world](./01-hello-world/) | Basic routing | None | ⭐ |
| [02-rest-api](./02-rest-api/) | REST API with middleware | None | ⭐⭐ |
| [03-generic-handlers](./03-generic-handlers/) | Type-safe Context[Req, Res] | None | ⭐⭐ |
| [04-middleware](./04-middleware/) | Custom middleware | None | ⭐⭐ |
| [05-validation](./05-validation/) | Request validation | None | ⭐⭐ |
| [06-openapi](./06-openapi/) | OpenAPI 3.1 generation | None | ⭐⭐⭐ |
| [07-sse-notifications](./07-sse-notifications/) | Server-Sent Events | plugins/stream | ⭐⭐⭐ |
| [08-websocket-chat](./08-websocket-chat/) | WebSocket chat | plugins/stream | ⭐⭐⭐ |
| [09-rest-api-with-db](./09-rest-api-with-db/) | REST API + Database | plugins/database | ⭐⭐⭐ |
| [10-production-boilerplate](./10-production-boilerplate/) | **Complete DDD Production App** | database + stream | ⭐⭐⭐⭐ |

## By Feature

### Core Features
- **Routing**: 01, 02
- **Type-safe Handlers**: 03
- **Middleware**: 02, 04
- **Validation**: 05
- **OpenAPI**: 06

### Real-time Communication (plugins/stream)
- **Server-Sent Events**: 07
- **WebSocket**: 08

### Database Integration (plugins/database)
- **CRUD Operations**: 09
- **Transactions**: 09 (batch endpoint)

## Getting Started

### Prerequisites

```bash
go version  # Requires Go 1.25+
```

### Running Examples

Each example is self-contained:

```bash
cd examples/01-hello-world
go run main.go
```

### Common Patterns

**1. Basic Router Setup**
```go
router := fursy.New()
router.GET("/", handler)
router.Run(":8080")
```

**2. With Database**
```go
import "github.com/coregx/fursy/plugins/database"

db := database.NewDB(sqlDB)
router.Use(database.Middleware(db))
```

**3. With SSE**
```go
import (
    "github.com/coregx/fursy/plugins/stream"
    "github.com/coregx/stream/sse"
)

hub := sse.NewHub[T]()
go hub.Run()
router.Use(stream.SSEHub(hub))
```

**4. With WebSocket**
```go
import (
    "github.com/coregx/fursy/plugins/stream"
    "github.com/coregx/stream/websocket"
)

hub := websocket.NewHub()
go hub.Run()
router.Use(stream.WebSocketHub(hub))
```

## Learning Path

**Beginner** (Start here!):
1. 01-hello-world - Basic routing
2. 02-rest-api - REST API basics
3. 03-generic-handlers - Type-safe handlers

**Intermediate**:
4. 04-middleware - Custom middleware
5. 05-validation - Request validation
6. 09-rest-api-with-db - Database integration

**Advanced**:
7. 07-sse-notifications - Real-time updates
8. 08-websocket-chat - Bidirectional communication
9. 06-openapi - API documentation

**Production**:
10. 10-production-boilerplate - **Complete DDD architecture** with JWT auth, real-time features, database, Docker support

## Additional Resources

- [fursy Documentation](../README.md)
- [plugins/stream Documentation](../plugins/stream/README.md)
- [plugins/database Documentation](../plugins/database/README.md)
- [stream library](https://github.com/coregx/stream)

## Contributing

Have a great example to share? PRs welcome!
