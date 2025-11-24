module github.com/coregx/fursy/plugins/stream

go 1.25.0

require (
	github.com/coregx/fursy v0.2.0
	github.com/coregx/stream v0.1.0
)

// Local development - replace with actual module paths.
replace github.com/coregx/fursy => ../..

replace github.com/coregx/stream => ../../../stream
