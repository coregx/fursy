module github.com/coregx/fursy

go 1.25.0

require (
	github.com/coregx/fursy/plugins/database v0.0.0-00010101000000-000000000000
	github.com/coregx/fursy/plugins/stream v0.0.0-00010101000000-000000000000
	github.com/coregx/stream v0.1.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	golang.org/x/time v0.14.0
	modernc.org/sqlite v1.40.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/sys v0.36.0 // indirect
	modernc.org/libc v1.66.10 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

replace github.com/coregx/stream => D:/projects/coregx/stream

replace github.com/coregx/fursy/plugins/stream => ./plugins/stream

replace github.com/coregx/fursy/plugins/database => ./plugins/database
