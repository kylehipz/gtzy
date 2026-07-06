// Package web embeds the built frontend (web/dist, produced by `npm run
// build`) so the Go binary can serve it directly. dist/.gitkeep keeps the
// directory present in source control so `go build` succeeds even before
// the frontend has been built.
package web

import "embed"

//go:embed all:dist
var DistFS embed.FS
