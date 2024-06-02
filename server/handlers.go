// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"bytes"
	reports "github.com/mdhender/ottomap/pkg/reports/domain"
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

func (s *Server) getDashboard() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(s.app.paths.templates, "dashboard.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		user, ok := s.app.policies.CurrentUser(r)
		if !ok {
			log.Printf("%s: %s: currentUser: not ok\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		} else if !user.IsAuthenticated {
			log.Printf("%s: %s: currentUser: not authenticated\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
	}
}

func (s *Server) getFeatures() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(s.app.paths.templates, "features.gohtml"),
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
	}
}

func (s *Server) getHero() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(s.app.paths.templates, "hero.gohtml"),
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
	}
}

// getIndex handles the index, static pages, and not found pages.
// Because of that quirk of Go where "/" matches most routes but
// not all routes, it gets called for the index and most pages that
// don't have a route assigned to them. For the index page, we want
// to serve our Hero page (getHero). For the other routes, we want
// to see if the route matches a file name in the static directory.
// If it does, we serve the file. If it doesn't, we serve the 404.
// The handleStaticFile function is responsible for doing all of this.
func (s *Server) getIndex() http.HandlerFunc {
	handleGetHero := s.getHero()
	handleGetStaticFiles := s.handleStaticFiles("/", s.app.paths.public, s.app.debug)
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		if r.URL.Path != "/" { // let the static file handler deal with it
			handleGetStaticFiles(w, r)
			return
		}
		// serve the hero page
		handleGetHero(w, r)
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
		user, ok := s.app.policies.CurrentUser(r)
		if ok && user.IsAuthenticated {
			log.Printf("%s: %s: user %q: ok && authenticate\n", r.Method, r.URL.Path, user.Id)
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		log.Printf("%s: %s: no active session, serving login form\n", r.Method, r.URL.Path)

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
	}
}

func (s *Server) postLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		// delete session and cookie (ignore errors if they don't exist)
		s.app.policies.DeleteSession(r)
		s.app.policies.DeleteCookie(w)

		handle := "ottomap"  // todo: post from form
		secret := "password" // todo: post from form

		// authenticate the user or return an error
		uid, ok := s.app.policies.Authenticate(handle, secret)
		if !ok {
			log.Printf("%s: %s: handle %q: secret %q: authenticate failed\n", r.Method, r.URL.Path, handle, secret)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// create a new session or return an error
		if _, ok := s.app.policies.CreateSession(w, uid); !ok {
			log.Printf("%s: %s: handle %q: secret %q: create session failed\n", r.Method, r.URL.Path, handle, secret)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("%s: %s: handle %q: secret %q: create session worked\n", r.Method, r.URL.Path, handle, secret)

		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

func (s *Server) getLogout() http.HandlerFunc {
	return s.postLogout()
}

func (s *Server) postLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// delete session and cookie (ignore errors if they don't exist)
		s.app.policies.DeleteSession(r)
		s.app.policies.DeleteCookie(w)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (s *Server) getReports() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(s.app.paths.templates, "reports.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		user, ok := s.app.policies.CurrentUser(r)
		if !ok {
			log.Printf("%s: %s: currentUser: not ok\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		} else if !user.IsAuthenticated {
			log.Printf("%s: %s: currentUser: not authenticated\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Parse the template file
		tmpl, err := template.ParseFiles(templateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var payload struct {
			Page struct {
				Title string
			}
			ReportList reports.ReportListing
		}
		payload.Page.Title = "Reports"

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
	}
}

// returns a handler that will serve a static file if one exists, otherwise return not found.
func (s *Server) handleStaticFiles(prefix, root string, debug bool) http.HandlerFunc {
	log.Println("[static] initializing")
	defer log.Println("[static] initialized")

	log.Printf("[static] strip: %q\n", prefix)
	log.Printf("[static]  root: %q\n", root)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

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
	}
}

func (s *Server) handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(s.app.version.String()))
	}
}

func (s *Server) apiGetLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//name := r.PathValue("name")
		//secret := r.PathValue("secret")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}
