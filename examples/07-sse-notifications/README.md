# SSE Notifications Example

Server-Sent Events (SSE) notification server using fursy router and stream plugin.

## Features

- **Real-time notifications** via Server-Sent Events
- **POST endpoint** to broadcast messages to all connected clients
- **Periodic updates** sent every 5 seconds
- **Type-safe events** with structured Notification type
- **Client tracking** - see how many clients are connected

## Running the Example

```bash
# Start the server
go run main.go
```

Server will start on http://localhost:8080

## Usage

### Listen for Events

Connect with curl:
```bash
curl -N http://localhost:8080/events
```

You'll receive:
1. Welcome message immediately
2. Periodic updates every 5 seconds
3. Any notifications sent via POST /notify

### Send Notification

In another terminal, send a notification:
```bash
curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{"type":"alert","message":"Important update!"}'
```

This will broadcast the message to **all connected clients**.

### Try Multiple Clients

Open 3 terminals and run `curl -N http://localhost:8080/events` in each.
Then send a notification - all 3 clients will receive it simultaneously!

## API Endpoints

### GET /events

Server-Sent Events endpoint. Clients connect here to receive notifications.

**Response:**
- Content-Type: `text/event-stream`
- Events in SSE format:
  ```
  data: {"type":"info","message":"Connected","time":"2025-01-18T..."}

  data: {"type":"alert","message":"Important update!","time":"2025-01-18T..."}
  ```

### POST /notify

Broadcast notification to all connected clients.

**Request:**
```json
{
  "type": "alert",
  "message": "Important update!"
}
```

**Response:**
```json
{
  "status": "sent",
  "clients": 3
}
```

## Event Types

The `Notification` struct supports different event types:

- **info**: Informational messages (periodic updates, welcome messages)
- **alert**: Important notifications that need attention
- **warning**: Warning messages
- **error**: Error notifications
- **success**: Success confirmations

## How It Works

1. **Hub Creation**: SSE Hub is created and started
2. **Middleware**: `stream.SSEHub(hub)` makes hub available to all handlers
3. **SSE Endpoint**: `/events` upgrades HTTP to SSE and registers client to hub
4. **Broadcasting**: Any handler can get hub and call `hub.BroadcastJSON(notification)`
5. **Automatic Cleanup**: Clients are unregistered when they disconnect

## Code Highlights

### SSE Upgrade

```go
return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
    hub.Register(conn)
    defer hub.Unregister(conn)

    // Send welcome message
    conn.SendJSON(Notification{
        Type:    "info",
        Message: "Connected",
        Time:    time.Now(),
    })

    // Wait until client disconnects
    <-conn.Done()
    return nil
})
```

### Broadcasting

```go
// Broadcast to all connected clients
hub.BroadcastJSON(Notification{
    Type:    "alert",
    Message: "System maintenance in 5 minutes",
    Time:    time.Now(),
})
```

## Production Considerations

1. **Authentication**: Add middleware to verify client credentials before SSE upgrade
2. **Rate Limiting**: Limit connections per IP to prevent abuse
3. **Heartbeats**: Periodic comments keep connections alive (already implemented in stream)
4. **Error Handling**: Log errors when clients disconnect unexpectedly
5. **Graceful Shutdown**: Close hub on server shutdown to notify all clients

## Next Steps

- Try the WebSocket chat example: `../08-websocket-chat`
- Read the [SSE Guide](https://github.com/coregx/stream/blob/main/docs/SSE_GUIDE.md)
- Explore [stream library](https://github.com/coregx/stream)
