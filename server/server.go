// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package server implements a web server for Otto.
package server

import (
	"fmt"
	"github.com/mdhender/ottomap/pkg/reports/dao"
	"github.com/mdhender/ottomap/pkg/simba"
	"github.com/mdhender/semver"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type Server struct {
	http.Server
	app struct {
		baseURL string
		paths   struct {
			root      string
			public    string
			css       string
			templates string
		}
		debug    bool
		dateFmt  string
		policies *simba.Agent
		stores   struct {
			// todo: maybe might should be the interfaces
			reports *reports.Store
		}
		version semver.Version
	}
	scheme  string
	host    string
	port    string
	mux     *http.ServeMux
	version semver.Version
}

// New returns a Server with default settings that are overridden by the provided options.
func New(options ...Option) (*Server, error) {
	s := &Server{
		scheme:  "http",
		host:    "localhost",
		port:    "3000",
		mux:     http.NewServeMux(), // default mux, no routes
		version: semver.Version{Major: 0, Minor: 1, Patch: 0},
	}

	s.IdleTimeout = 10 * time.Second
	s.ReadTimeout = 5 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxHeaderBytes = 1 << 20

	s.app.paths.root = "."
	s.app.paths.public = filepath.Join(s.app.paths.root, "..", "public")
	s.app.paths.css = filepath.Join(s.app.paths.public, "css")
	s.app.paths.templates = filepath.Join(s.app.paths.root, "..", "templates")
	s.app.dateFmt = "2006-01-02"

	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	s.app.baseURL = s.BaseURL()

	if err := isdir(s.app.paths.root); err != nil {
		return nil, err
	} else {
		log.Printf("server: root      is %s\n", s.app.paths.root)
	}
	if err := isdir(s.app.paths.public); err != nil {
		return nil, err
	} else {
		log.Printf("server: public    is %s\n", s.app.paths.public)
	}
	if err := isdir(s.app.paths.css); err != nil {
		return nil, err
	} else {
		log.Printf("server: css       is %s\n", s.app.paths.css)
	}
	if err := isdir(s.app.paths.templates); err != nil {
		return nil, err
	} else {
		log.Printf("server: templates is %s\n", s.app.paths.templates)
	}

	if s.app.policies == nil {
		return nil, fmt.Errorf("missing policies agent")
	} else if s.app.stores.reports == nil {
		return nil, fmt.Errorf("missing reports store")
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
	log.Printf("serve: %s%s\n", s.app.baseURL, "/")
	log.Printf("serve: %s%s\n", s.app.baseURL, "/login")
	log.Printf("serve: %s%s\n", s.app.baseURL, "/logout")
	log.Printf("serve: %s%s\n", s.app.baseURL, "/dashboard")
	log.Printf("serve: %s%s\n", s.app.baseURL, "/reports")
	log.Printf("serve: %s%s\n", s.app.baseURL, "/reports/0991")
	log.Printf("serve: %s%s\n", s.app.baseURL, "/api/version")
}
