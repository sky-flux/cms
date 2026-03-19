// Package cms is the root package of the Sky Flux CMS module.
// Its sole responsibility is to expose embedded file systems for the
// compiled console SPA and web static assets, so cmd/cms/main.go can
// import them via a single import path.
//
// Build: ensure console/dist is populated before `go build`:
//   cd console && bun run build
//
// Development: console/dist/.gitkeep is committed so `go:embed` compiles.
// The Chi router detects --dev flag and proxies /console/* to Vite :3000.
package cms

import "embed"

// ConsoleFS holds the React admin SPA production build.
// In development, the Go server proxies /console/* to Vite.
//
//go:embed all:console/dist
var ConsoleFS embed.FS

// WebStaticFS holds compiled Tailwind CSS and HTMX for the public site.
//
//go:embed all:web/static
var WebStaticFS embed.FS
