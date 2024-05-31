// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"net/http"
)

// todo: implement https://go.dev/blog/routing-enhancements

func (s *Server) Routes() http.Handler {
	// add our public routes
	for _, route := range []struct {
		pattern string
		handler http.HandlerFunc
	}{
		{"GET /", s.getIndex()},
		{"GET /features", s.getFeatures()},
		{"GET /login", s.getLogin()},
		{"POST /login", s.postLogin()},
		{"GET /logout", s.getLogout()},
		{"POST /logout", s.postLogout()},
		{"GET /api/version", s.handleVersion()},
		{"GET /api/login/{name}/{secret}", s.apiGetLogin()},
	} {
		s.mux.HandleFunc(route.pattern, route.handler)
	}

	// add our protected routes
	for _, route := range []struct {
		pattern string
		handler http.HandlerFunc
	}{
		{"GET /dashboard", s.getDashboard()},
		{"GET /reports", s.getReports()},
	} {
		s.mux.HandleFunc(route.pattern, s.addSession(s.mustAuthenticate(route.handler)))
	}

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
