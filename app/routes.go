// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package app

import (
	reports "github.com/mdhender/ottomap/pkg/reports/handlers"
	turns "github.com/mdhender/ottomap/pkg/turns/handlers"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// todo: implement https://go.dev/blog/routing-enhancements

func (a *App) Routes() (*http.ServeMux, error) {
	mux := http.NewServeMux() // default mux, no routes

	mux.HandleFunc("GET /", handleGetHero02(a.paths.templates, a.policies))
	mux.HandleFunc("GET /dashboard", handleGetDashboard(a.paths.templates, a.policies, a.db))
	mux.HandleFunc("GET /features", a.getFeatures())
	mux.HandleFunc("GET /login", a.getLogin())
	mux.HandleFunc("POST /login", a.postLogin())
	mux.HandleFunc("GET /logout", a.getLogout())
	mux.HandleFunc("POST /logout", a.postLogout())
	mux.Handle("GET /reports", reports.HandleGetReportsListing(a.paths.templates, a.policies, a.db))
	mux.Handle("GET /reports/{reportId}", reports.HandleGetReportsListing(a.paths.templates, a.policies, a.db))
	mux.Handle("GET /turns", turns.HandleGetListing(a.paths.templates, a.policies, a.db))
	mux.Handle("GET /turns/{turnId}", turns.HandleGetDetail(a.paths.templates, a.policies, nil))

	mux.HandleFunc("GET /api/version", a.handleVersion())
	mux.HandleFunc("GET /api/login/{name}/{secret}", a.apiGetLogin())

	// walk the public directory and add routes to serve static files
	validExtensions := map[string]bool{
		".css":    true,
		".html":   true,
		".ico":    true,
		".jpg":    true,
		".js":     true,
		".png":    true,
		".robots": true,
		".svg":    true,
	}
	if err := filepath.WalkDir(a.paths.public, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		} else if !validExtensions[filepath.Ext(path)] {
			return nil
		} else if strings.HasPrefix(filepath.Base(path), ".") { // avoid serving .dotfiles
			return nil
		}
		route := "GET " + strings.TrimPrefix(path, a.paths.public)
		log.Printf("app: public    adding route for %s\n", path)
		log.Printf("app: path  %q\n", path)
		log.Printf("app: route %q\n", route)
		mux.Handle(route, handleStaticFiles("", a.paths.public, false))
		return nil
	}); err != nil {
		return nil, err
	}

	return mux, nil
}
