# WebSocket Chat Example

Real-time chat application using fursy router and WebSocket support from stream plugin.

## Features

- **Real-time messaging** via WebSocket
- **Broadcast messages** to all connected clients
- **System notifications** for join/leave events
- **Type-safe messages** with structured ChatMessage type
- **Health check endpoint** showing connected clients count

## Running the Example

```bash
# Start the server
go run main.go
```

Server will start on http://localhost:8080

## Usage

### Using wscat (CLI)

Install wscat:
```bash
npm install -g wscat
```

Connect:
```bash
wscat -c ws://localhost:8080/ws
```

Send messages as JSON:
```json
{"user":"Alice","message":"Hello, everyone!"}
```

### Using Browser

Create `client.html`:
```html
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Chat</title>
</head>
<body>
    <h1>WebSocket Chat</h1>
    <div id="messages"></div>
    <input type="text" id="user" placeholder="Your name" value="User1">
    <input type="text" id="message" placeholder="Type message...">
    <button onclick="sendMessage()">Send</button>

    <script>
        const ws = new WebSocket('ws://localhost:8080/ws');

        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            const div = document.getElementById('messages');
            div.innerHTML += `<p><strong>${msg.user}:</strong> ${msg.message} <small>(${new Date(msg.time).toLocaleTimeString()})</small></p>`;
        };

        function sendMessage() {
            const user = document.getElementById('user').value;
            const message = document.getElementById('message').value;
            ws.send(JSON.stringify({user, message}));
            document.getElementById('message').value = '';
        }
    </script>
</body>
</html>
```

Open `client.html` in multiple browsers to test multi-client chat!

### Try Multiple Clients

1. Open 3 terminals
2. Run `wscat -c ws://localhost:8080/ws` in each
3. Type messages in any terminal - all clients will receive them!

## API Endpoints

### GET /ws

WebSocket endpoint. Clients connect here for real-time chat.

**Connection:**
```
ws://localhost:8080/ws
```

**Message Format (send):**
```json
{
  "user": "Alice",
  "message": "Hello, everyone!"
}
```

**Message Format (receive):**
```json
{
  "user": "Alice",
  "message": "Hello, everyone!",
  "time": "2025-01-18T12:34:56.789Z"
}
```

### GET /health

Health check endpoint showing server status and connected clients.

**Response:**
```json
{
  "status": "ok",
  "clients": 3,
  "time": "2025-01-18T12:34:56.789Z"
}
```

## How It Works

1. **Hub Creation**: WebSocket Hub is created and started
2. **Middleware**: `stream.WebSocketHub(hub)` makes hub available to all handlers
3. **WebSocket Endpoint**: `/ws` upgrades HTTP to WebSocket and registers client to hub
4. **Message Broadcasting**: Messages are read from client and broadcast to all via hub
5. **Automatic Cleanup**: Clients are unregistered when they disconnect

## Code Highlights

### WebSocket Upgrade

```go
return stream.WebSocketUpgrade(c, func(conn *websocket.Conn) error {
    hub.Register(conn)
    defer hub.Unregister(conn)

    // Send welcome message
    conn.WriteJSON(ChatMessage{
        User:    "System",
        Message: "Welcome to the chat!",
        Time:    time.Now(),
    })

    // Read loop
    for {
        var msg ChatMessage
        if err := conn.ReadJSON(&msg); err != nil {
            return err // Client disconnected
        }

        msg.Time = time.Now()
        hub.BroadcastJSON(msg) // Broadcast to all
    }
}, nil)
```

### Broadcasting

```go
// Broadcast to all connected clients
hub.BroadcastJSON(ChatMessage{
    User:    "System",
    Message: "Server maintenance in 5 minutes",
    Time:    time.Now(),
})
```

## Message Types

The `ChatMessage` struct supports different message scenarios:

### User Message
```json
{
  "user": "Alice",
  "message": "Hello, everyone!",
  "time": "2025-01-18T12:34:56.789Z"
}
```

### System Message
```json
{
  "user": "System",
  "message": "A new user has joined the chat",
  "time": "2025-01-18T12:34:56.789Z"
}
```

## Production Considerations

1. **Authentication**: Verify user identity before WebSocket upgrade
2. **Rate Limiting**: Limit messages per client to prevent spam
3. **Message Validation**: Sanitize/validate messages before broadcasting
4. **User Management**: Track connected users and their metadata
5. **Persistence**: Store chat history in database
6. **Room Support**: Create separate channels/rooms for different topics
7. **Typing Indicators**: Send typing events for better UX
8. **Read Receipts**: Track message delivery and read status

## Advanced Features to Add

### Private Messages

```go
// Send to specific client only
targetConn := findUserConnection(targetUser)
targetConn.WriteJSON(privateMsg)
```

### Chat Rooms

```go
// Create multiple hubs for different rooms
rooms := map[string]*websocket.Hub{
    "general": websocket.NewHub(),
    "tech":    websocket.NewHub(),
    "random":  websocket.NewHub(),
}
```

### Presence Detection

```go
// Heartbeat mechanism
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        if err := conn.Ping(); err != nil {
            return // Client disconnected
        }
    }
}()
```

## Next Steps

- Try the SSE notifications example: `../07-sse-notifications`
- Read the [WebSocket Guide](https://github.com/coregx/stream/blob/main/docs/WEBSOCKET_GUIDE.md)
- Explore [stream library](https://github.com/coregx/stream)
