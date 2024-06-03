// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package server implements a web server for Otto.
package server

import (
	"fmt"
	"github.com/mdhender/ottomap/app"
	"log"
	"net/http"
	"time"
)

type Server struct {
	http.Server
	app    *app.App
	scheme string
	host   string
	port   string
	mux    *http.ServeMux
}

// New returns a Server with default settings that are overridden by the provided options.
func New(options ...Option) (*Server, error) {
	s := &Server{
		scheme: "http",
		host:   "localhost",
		port:   "3000",
		mux:    http.NewServeMux(), // default mux, no routes
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
