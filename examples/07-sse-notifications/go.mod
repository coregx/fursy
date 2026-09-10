module sse-notifications

go 1.27

require (
	github.com/coregx/fursy v0.5.3
	github.com/coregx/fursy/plugins/stream v0.0.0
	github.com/coregx/stream v0.1.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	golang.org/x/time v0.16.0 // indirect
)

replace (
	github.com/coregx/fursy => ../..
	github.com/coregx/fursy/plugins/stream => ../../plugins/stream
)

replace github.com/coregx/fursy/plugins/database => ../../plugins/database
