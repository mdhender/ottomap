// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"github.com/mdhender/ottomap/sessions"
	"net/http"
)

func (s *Server) mustAuthenticate(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// call the next handler if the user is authenticated, otherwise return a forbidden status
		if !sessions.User(ctx).IsAuthenticated {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
