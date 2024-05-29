// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package server implements a web server for Otto.
package server

import (
	"context"
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
		scheme string
		host   string
		port   string
		paths  struct {
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
	s.app.paths.public = filepath.Join(s.app.paths.root, "public")
	s.app.paths.css = filepath.Join(s.app.paths.public, "css")
	s.app.paths.templates = filepath.Join(s.app.paths.root, "templates")
	s.app.dateFmt = "2006-01-02"

	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, err
		}
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
	log.Printf("serve: %s://%s%s\n", s.app.scheme, s.Addr, "/api/version")
	for _, da := range s.users.store.TheSecrets() {
		log.Printf("serve: %s://%s/login/%s/%s\n", s.app.scheme, s.Addr, da[0], da[1])
	}
	log.Printf("serve: %s://%s%s\n", s.app.scheme, s.Addr, "/logout")
}

func (s *Server) currentUser(ctx context.Context) users.User {
	return sessions.User(ctx)
}
