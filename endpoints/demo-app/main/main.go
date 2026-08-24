// Command demo-app runs the OpenNHP Demo application.
//
// Usage:
//
//	demo-app --config etc/config.toml
//
// Configuration is a single TOML file; see config.go for the schema.
// The Demo App listens on cfg.ListenAddr (default :8088) and serves the
// static assets embedded in the binary.
package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"

	demoapp "github.com/OpenNHP/opennhp/endpoints/demo-app"
)

// //go:embed all:web pulls templates + static assets into the binary so
// the deployment artifact is a single file. The embed glob will fail at
// build time if the web/ directory is missing; we ship a placeholder
// index.html alongside this command so a fresh checkout always builds.
//
//go:embed all:web
var webFS embed.FS

func main() {
	cfgPath := flag.String("config", "etc/config.toml", "path to TOML config")
	flag.Parse()

	cfg, err := demoapp.LoadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(2)
	}

	db, err := demoapp.Open(cfg.DbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db error: %v\n", err)
		os.Exit(2)
	}
	defer db.Close()

	mailer, err := demoapp.NewMailer(cfg.SMTP)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mailer error: %v\n", err)
		os.Exit(2)
	}

	store := demoapp.NewCookieStore([]byte(cfg.SessionSecret))

	app, err := demoapp.New(cfg, db, mailer, store, webFS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init error: %v\n", err)
		os.Exit(2)
	}
	app.SetupRoutes()

	log.Printf("demo-app: listening on %s (db=%s, smtp=%s)", cfg.ListenAddr, cfg.DbPath, cfg.SMTP.Mode)
	if err := app.Engine.Run(cfg.ListenAddr); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
