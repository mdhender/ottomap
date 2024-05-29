// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package server implements a web server for Otto.
package server

import (
	"context"
	"fmt"
	"github.com/mdhender/ottomap/authz"
	"github.com/mdhender/ottomap/sessions"
	"github.com/mdhender/ottomap/users"
	"github.com/mdhender/ottomap/way"
	"github.com/mdhender/semver"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type Server struct {
	http.Server
	app struct {
		scheme  string
		host    string
		port    string
		baseURL string
		paths   struct {
			root      string
			public    string
			css       string
			templates string
		}
		debug   bool
		dateFmt string
	}
	sessions struct {
		path    string
		store   *sessions.Store
		manager *sessions.Manager
		cookies struct {
			name   string
			secure bool
		}
	}
	users struct {
		path  string
		store *users.Store
	}
	auth struct {
		secret  string
		manager *authz.Factory
		ttl     time.Duration
	}
	router  *way.Router
	version semver.Version
}

// New returns a Server with default settings that are overridden by the provided options.
func New(options ...Option) (*Server, error) {
	s := &Server{
		router:  way.NewRouter(), // default router, no routes
		version: semver.Version{Major: 0, Minor: 1, Patch: 0},
	}

	s.IdleTimeout = 10 * time.Second
	s.ReadTimeout = 5 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxHeaderBytes = 1 << 20

	s.app.scheme = "http"
	s.app.host = "localhost"
	s.app.port = "3000"
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

	s.app.baseURL = fmt.Sprintf("%s://%s", s.app.scheme, s.Addr)

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
	if err := isfile(s.users.path); err != nil {
		return nil, err
	} else {
		log.Printf("server: user     store is %s\n", s.users.path)
	}
	if err := isfile(s.sessions.path); err != nil {
		return nil, err
	} else {
		log.Printf("server: sessions store is %s\n", s.sessions.path)
	}

	var err error
	s.users.store, err = users.New(s.users.path)
	if err != nil {
		return nil, err
	}

	s.sessions.store, err = sessions.NewStore(s.sessions.path, s.users.store)
	if err != nil {
		return nil, err
	}
	s.sessions.manager, err = sessions.NewManager(s.sessions.cookies.name, s.sessions.store, s.users.store)
	if err != nil {
		return nil, err
	}

	s.auth.manager, err = authz.New("ottomap", s.auth.secret, 2*7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) ShowMeSomeRoutes() {
	log.Printf("serve: %s%s\n", s.app.baseURL, "/")
	log.Printf("serve: %s%s\n", s.app.baseURL, "/login")
	for _, da := range s.users.store.TheSecrets() {
		log.Printf("serve: %s/login/%s/%s\n", s.app.baseURL, da[0], da[1])
	}
	log.Printf("serve: %s%s\n", s.app.baseURL, "/logout")
	log.Printf("serve: %s%s\n", s.app.baseURL, "/api/version")
}

func (s *Server) currentUser(ctx context.Context) users.User {
	return sessions.User(ctx)
}
