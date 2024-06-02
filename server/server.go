// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package server implements a web server for Otto.
package server

import (
	"fmt"
	"github.com/mdhender/ottomap/app"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Server struct {
	http.Server
	app    *app.App
	scheme string
	host   string
	port   string
	mux    *http.ServeMux
	public string // path to serve public (static) files from
}

// New returns a Server with default settings that are overridden by the provided options.
func New(options ...Option) (*Server, error) {
	s := &Server{
		scheme: "http",
		host:   "localhost",
		port:   "3000",
		mux:    http.NewServeMux(), // default mux, no routes
		public: "public",
	}

	s.IdleTimeout = 10 * time.Second
	s.ReadTimeout = 5 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxHeaderBytes = 1 << 20

	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	if err := isdir(s.public); err != nil {
		return nil, err
	}
	log.Printf("server: public    is %s\n", s.public)

	// the above and beyond public file handler finds all the files in the public directory,
	// and adds routes to serve them as static files.
	//

	// walk the public directory and add routes to serve files
	validExtensions := map[string]bool{
		".css":  true,
		".html": true,
		".ico":  true,
		".jpg":  true, ".js": true,
		".png":    true,
		".robots": true,
		".svg":    true,
	}
	if err := filepath.WalkDir(s.public, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		} else if !validExtensions[filepath.Ext(path)] {
			return nil
		} else if strings.HasPrefix(filepath.Base(path), ".") { // avoid serving .dotfiles
			return nil
		}
		route := "GET " + strings.TrimPrefix(path, s.public)
		log.Printf("server: public    adding route for %s\n", path)
		log.Printf("server: path  %q\n", path)
		log.Printf("server: route %q\n", route)
		s.mux.Handle(route, s.handleStaticFiles("", s.public, false))
		return nil
	}); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) BaseURL() string {
	return fmt.Sprintf("%s://%s", s.scheme, s.Addr)
}

func (s *Server) Router() http.Handler {
	return s.mux
}

func (s *Server) ShowMeSomeRoutes() {
	log.Printf("serve: %s%s\n", s.BaseURL(), "/")
	log.Printf("serve: %s%s\n", s.BaseURL(), "/login")
	log.Printf("serve: %s%s\n", s.BaseURL(), "/logout")
	log.Printf("serve: %s%s\n", s.BaseURL(), "/dashboard")
	log.Printf("serve: %s%s\n", s.BaseURL(), "/reports")
	log.Printf("serve: %s%s\n", s.BaseURL(), "/reports/0991")
	log.Printf("serve: %s%s\n", s.BaseURL(), "/api/version")
}
