//go:build webdist

// This file is only compiled when the `webdist` build tag is set, which
// happens automatically when the Dockerfile.demoapp stage 2 builds the
// binary with web/dist/ present. The embed.FS below ships the bundled SPA
// inside the demoapp binary so the runtime image needs nothing but the
// binary itself.
package main

import (
	"embed"
	"io/fs"
)

//go:embed web/dist
var webDistFS embed.FS

func init() {
	sub, err := fs.Sub(webDistFS, "web/dist")
	if err == nil {
		webFS = sub
	}
}
