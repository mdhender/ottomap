// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"net/http"
)

// todo: implement https://go.dev/blog/routing-enhancements

func (s *Server) Routes() http.Handler {
	s.mux.HandleFunc("GET /", s.getIndex())
	s.mux.HandleFunc("GET /dashboard", s.getDashboard())
	s.mux.HandleFunc("GET /features", s.getFeatures())
	s.mux.HandleFunc("GET /login", s.getLogin())
	s.mux.HandleFunc("POST /login", s.postLogin())
	s.mux.HandleFunc("GET /logout", s.getLogout())
	s.mux.HandleFunc("POST /logout", s.postLogout())
	s.mux.HandleFunc("GET /reports", s.getReports())
	s.mux.Handle("GET /reports/0991", handleReportsListing(s.app.paths.templates, s.app.policies, s.app.stores.reports))

	s.mux.HandleFunc("GET /api/version", s.handleVersion())
	s.mux.HandleFunc("GET /api/login/{name}/{secret}", s.apiGetLogin())

	// add our not found handler (it will serve public files if they exist)
	s.mux.Handle("/", s.handleStaticFiles("/", s.app.paths.public, s.app.debug))

	return s.mux
}

// responseWriterWrapper is used to capture the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriterWrapper) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
