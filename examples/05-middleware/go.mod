module example.com/middleware

go 1.25.0

replace github.com/coregx/fursy => ../..

require (
	github.com/coregx/fursy v0.1.0
	github.com/golang-jwt/jwt/v5 v5.3.0
)

require golang.org/x/time v0.14.0 // indirect
