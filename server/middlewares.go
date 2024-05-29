// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"github.com/mdhender/ottomap/sessions"
	"log"
	"net/http"
)

// addSession fetches the session from the cookie.
// if the session is valid, the user data is added to the request context.
// otherwise, the anonymous (unauthenticated) user is added.
func (s *Server) addSession(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessId, ok := s.sessions.manager.GetCookie(r)
		if !ok {
			log.Printf("%s: %s: missing session cookie\n", r.Method, r.URL.Path)
			ctx := s.sessions.manager.AddUser(r.Context(), s.users.store.AnonymousUser())
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		sess, ok := s.sessions.manager.GetSession(sessId)
		if !ok {
			log.Printf("%s: %s: invalid session %q\n", r.Method, r.URL.Path, sessId)
			ctx := s.sessions.manager.AddUser(r.Context(), s.users.store.AnonymousUser())
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		log.Printf("%s: %s: adding user %q\n", r.Method, r.URL.Path, sess.User.Handle)

		ctx := s.sessions.manager.AddUser(r.Context(), sess.User)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// call the next handler if the user is authenticated, otherwise return a forbidden status
func (s *Server) mustAuthenticate(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := sessions.User(r.Context())
		if !user.IsAuthenticated {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
