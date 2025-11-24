module websocket-chat

go 1.25.0

require (
	github.com/coregx/fursy v0.2.0
	github.com/coregx/fursy/plugins/stream v0.0.0
	github.com/coregx/stream v0.1.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)

replace (
	github.com/coregx/fursy => ../..
	github.com/coregx/fursy/plugins/stream => ../../plugins/stream
	github.com/coregx/stream => D:/projects/coregx/stream
)

replace github.com/coregx/fursy/plugins/database => ../../plugins/database
