// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"net/http"
)

func (s *Server) Routes() http.Handler {
	// add our public routes
	for _, route := range []struct {
		pattern string
		method  string
		handler http.HandlerFunc
	}{
		{"/", "GET", s.handleIndex()},
		{"/api/version", "GET", s.handleVersion()},
		{"/login/:name/:secret", "GET", s.handleLogin()},
		{"/logout", "GET", s.handleLogout()},
		{"/logout", "POST", s.handleLogout()},
	} {
		s.router.HandleFunc(route.method, route.pattern, route.handler)
	}

	// add our protected routes
	for _, route := range []struct {
		pattern string
		method  string
		handler http.HandlerFunc
	}{} {
		s.router.HandleFunc(route.method, route.pattern, s.mustAuthenticate(route.handler))
	}
	// add our not found handler (it will serve public files if they exist)
	s.router.NotFound = s.handleStaticFiles("/", s.app.paths.public, s.app.debug)

	return s.router
}
