module example.com/production-boilerplate

go 1.25.0

require (
	github.com/coregx/fursy v0.3.0
	github.com/coregx/stream v0.1.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.32.0
	modernc.org/sqlite v1.40.1
)

require (
	github.com/coregx/fursy/plugins/database v0.0.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	modernc.org/libc v1.66.10 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

replace github.com/coregx/fursy => ../..

replace github.com/coregx/fursy/plugins/database => ../../plugins/database

replace github.com/coregx/stream => ../../../stream
