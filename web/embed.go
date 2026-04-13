package web

import "embed"

// Dist is the production Vite bundle (build with npm run build:embed in this directory).
//
//go:embed all:dist
var Dist embed.FS
