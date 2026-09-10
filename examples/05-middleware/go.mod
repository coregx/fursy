module example.com/middleware

go 1.27

replace github.com/coregx/fursy => ../..

require (
	github.com/coregx/fursy v0.1.0
	github.com/golang-jwt/jwt/v5 v5.3.1
)

require golang.org/x/time v0.16.0 // indirect
