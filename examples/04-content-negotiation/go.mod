module example.com/content-negotiation

go 1.27

replace github.com/coregx/fursy => ../..

require github.com/coregx/fursy v0.5.3

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	golang.org/x/time v0.16.0 // indirect
)
