// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	http.Server
	paths struct {
		data      string
		public    string
		templates string
	}
	scheme string
	host   string
	port   string
	mux    *http.ServeMux
	router http.Handler
}

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

	for _, option := range options {
		if err := option(s); err != nil {
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
	log.Printf("serve: %s%s\n", s.BaseURL(), "/tn3")
	for _, turn := range []string{"0899-12", "0900-01", "0900-02"} {
		log.Printf("serve: %s/tn3/%s\n", s.BaseURL(), turn)
		log.Printf("serve: %s/tn3/%s/0138\n", s.BaseURL(), turn)
		log.Printf("serve: %s/tn3/%s/0138/report\n", s.BaseURL(), turn)
		log.Printf("serve: %s/tn3/%s/0138/map\n", s.BaseURL(), turn)
	}
	log.Printf("serve: %s%s\n", s.BaseURL(), "/login")
	log.Printf("serve: %s%s\n", s.BaseURL(), "/logout")
}
