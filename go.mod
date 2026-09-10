module github.com/coregx/fursy

go 1.27

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	golang.org/x/time v0.16.0
)

replace github.com/coregx/fursy/plugins/stream => ./plugins/stream

replace github.com/coregx/fursy/plugins/database => ./plugins/database
