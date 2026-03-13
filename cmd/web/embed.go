package web

import "embed"

// webFS holds the embedded shell SPA. In production builds, this contains
// the Vite build output from web/dist/. During development, use --dev
// to proxy to Vite's dev server instead.
//
//go:embed all:placeholder
var webFS embed.FS
