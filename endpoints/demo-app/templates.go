// Template loading for the Demo App.
//
// The embedded filesystem is supplied by main/main.go via the
// `webFS` parameter, which embeds the entire `web/` tree. Keeping the
// embed directive in main lets the package compile even when the
// `web/` directory hasn't been populated yet (e.g., for `go vet` in CI).
package demoapp

import (
	"html/template"
	"io/fs"
)

// staticFS is the static asset FS passed into App; it's set by
// NewFromFS / New.
var staticFS fs.FS

// parseTemplates loads every *.html in web/templates from the supplied
// fs.FS (rooted at the `web/` directory). A single template set is
// shared across handlers — each page is referenced by its base
// filename (e.g. "login.html") so page-specific data doesn't leak
// between templates.
func parseTemplates(webFS fs.FS) (*template.Template, error) {
	funcs := template.FuncMap{
		"safeURL": func(s string) template.URL { return template.URL(s) },
	}
	t := template.New("").Funcs(funcs)
	templatesSub, err := fs.Sub(webFS, "templates")
	if err != nil {
		return nil, err
	}
	staticSub, err := fs.Sub(webFS, "static")
	if err == nil {
		staticFS = staticSub
	}
	entries, err := fs.ReadDir(templatesSub, ".")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !isHTML(e.Name()) {
			continue
		}
		content, err := fs.ReadFile(templatesSub, e.Name())
		if err != nil {
			return nil, err
		}
		if _, err := t.New(e.Name()).Parse(string(content)); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func isHTML(name string) bool {
	if len(name) < 5 {
		return false
	}
	return name[len(name)-5:] == ".html"
}
