// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package app

import (
	"bytes"
	reports "github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/simba"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func (a *App) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		_, _ = w.Write([]byte(`index`))
	}
}

func (a *App) getDashboard() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(a.paths.templates, "dashboard.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		user, ok := a.policies.CurrentUser(r)
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

func (a *App) getFeatures() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(a.paths.templates, "features.gohtml"),
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

func (a *App) getHero() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(a.paths.templates, "hero.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		log.Printf("%s: %s: root      %q\n", r.Method, r.URL.Path, a.paths.root)
		log.Printf("%s: %s: templates %q\n", r.Method, r.URL.Path, a.paths.templates)

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

func (a *App) getLanding() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(a.paths.templates, "landing.gohtml"),
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

func (a *App) getLogin() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(a.paths.templates, "login.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		user, ok := a.policies.CurrentUser(r)
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

func (a *App) postLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		// delete session and cookie (ignore errors if they don't exist)
		a.policies.DeleteSession(r)
		a.policies.DeleteCookie(w)

		handle := "ottomap"  // todo: post from form
		secret := "password" // todo: post from form

		// authenticate the user or return an error
		uid, ok := a.policies.Authenticate(handle, secret)
		if !ok {
			log.Printf("%s: %s: handle %q: secret %q: authenticate failed\n", r.Method, r.URL.Path, handle, secret)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// create a new session or return an error
		if _, ok := a.policies.CreateSession(w, uid); !ok {
			log.Printf("%s: %s: handle %q: secret %q: create session failed\n", r.Method, r.URL.Path, handle, secret)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("%s: %s: handle %q: secret %q: create session worked\n", r.Method, r.URL.Path, handle, secret)

		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

func (a *App) getLogout() http.HandlerFunc {
	return a.postLogout()
}

func (a *App) postLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// delete session and cookie (ignore errors if they don't exist)
		a.policies.DeleteSession(r)
		a.policies.DeleteCookie(w)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (a *App) getReports() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(a.paths.templates, "reports.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		user, ok := a.policies.CurrentUser(r)
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

func (a *App) handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(a.version.String()))
	}
}

func (a *App) apiGetLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//name := r.PathValue("name")
		//secret := r.PathValue("secret")
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

type ReportListingPage struct {
	Page struct {
		Title string
	}
	ReportList reports.ReportListing
}

type ReportListingRepository interface {
	AllReports(authorize func(reports.Report) bool) (reports.ReportListing, error)
}

func handleReportsListing(templatesPath string, a *simba.Agent, rlr ReportListingRepository) http.Handler {
	templateFiles := []string{
		filepath.Join(templatesPath, "reports.gohtml"),
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		user, ok := a.CurrentUser(r)
		if !ok {
			log.Printf("%s: %s: currentUser: not ok\n", r.Method, r.URL.Path)
			http.Redirect(w, r, "/logout", http.StatusFound)
			return
		} else if !user.IsAuthenticated {
			log.Printf("%s: %s: currentUser: not authenticated\n", r.Method, r.URL.Path)
			http.Redirect(w, r, "/logout", http.StatusFound)
			return
		}

		// Parse the template file
		tmpl, err := template.ParseFiles(templateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// the logic for the "service" should be a bunch of simple calls to the repository.
		var result ReportListingPage
		result.ReportList, err = rlr.AllReports(a.UserReportsFilter(user.Id))

		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
		buf := &bytes.Buffer{}

		// execute the template with our payload
		err = tmpl.Execute(buf, result)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	})
}
