// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"bytes"
	"github.com/mdhender/ottomap/sessions"
	"github.com/mdhender/ottomap/way"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		_, _ = w.Write([]byte(`index`))
	}
}

func (s *Server) getLanding() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(s.app.paths.templates, "landing.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		// Parse the template file
		tmpl, err := template.ParseFiles(templateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var payload struct{}

		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
		buf := &bytes.Buffer{}

		// execute the template with our payload
		err = tmpl.Execute(buf, payload)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
		_, _ = w.Write([]byte(``))
	}
}

func (s *Server) getLogin() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(s.app.paths.templates, "login.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		user := sessions.User(r.Context())
		if user.IsAuthenticated {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Parse the template file
		tmpl, err := template.ParseFiles(templateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var payload struct{}

		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
		buf := &bytes.Buffer{}

		// execute the template with our payload
		err = tmpl.Execute(buf, payload)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
		_, _ = w.Write([]byte(``))
	}
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// if there's a session cookie, delete it and the session
		if id, ok := s.sessions.manager.GetCookie(r); ok {
			// delete the session
			s.sessions.manager.DeleteSession(id)
			// delete the session cookie
			s.sessions.manager.DeleteCookie(w)
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<p>You have been logged out and your session has been invalidated. Please close your browser and re-open to log in again."))
	}
}

// returns a handler that will serve a static file if one exists, otherwise return not found.
func (s *Server) handleStaticFiles(prefix, root string, debug bool) http.Handler {
	log.Println("[static] initializing")
	defer log.Println("[static] initialized")

	log.Printf("[static] strip: %q\n", prefix)
	log.Printf("[static]  root: %q\n", root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if r.Method != "GET" || !sessions.User(ctx).IsAuthenticated {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		file := filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))
		if debug {
			log.Printf("[static] %q\n", file)
		}

		stat, err := os.Stat(file)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// only serve regular files, never directories or directory listings.
		if stat.IsDir() || !stat.Mode().IsRegular() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// pretty sure that we have a regular file at this point.
		rdr, err := os.Open(file)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		defer func(r io.ReadCloser) {
			_ = r.Close()
		}(rdr)

		// let Go serve the file. it does magic things like content-type, etc.
		http.ServeContent(w, r, file, stat.ModTime(), rdr)
	})
}

func (s *Server) handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(s.version.String()))
	}
}

func (s *Server) apiGetLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := way.Param(r.Context(), "name")
		secret := way.Param(r.Context(), "secret")

		// authenticate the user or return an error
		user, ok := s.users.store.Authenticate(name, secret)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// create a new session or return an error
		sessionId, ok := s.sessions.manager.CreateSession(user.Id)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// create the session cookie or return an error
		ok = s.sessions.manager.AddCookie(w, sessionId)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// add the user to the request context or return an error
		ctx := s.sessions.manager.AddUser(r.Context(), user)
		if ctx == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r.WithContext(ctx), "/", http.StatusFound)
	}
}
