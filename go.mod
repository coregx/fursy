module github.com/coregx/fursy

go 1.25.0

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	golang.org/x/time v0.14.0
)

replace github.com/coregx/fursy/plugins/stream => ./plugins/stream

replace github.com/coregx/fursy/plugins/database => ./plugins/database
