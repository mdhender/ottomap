// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package app

import (
	"bytes"
	"fmt"
	reports "github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/simba"
	turns "github.com/mdhender/ottomap/pkg/turns/domain"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
			ReportList reports.Listing
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

type DashboardPage struct {
	Page struct {
		Title string
	}
	Reports []DashboardTurnLine
}
type DashboardTurnLine struct {
	Turn    string
	Reports []DashboardReportLine
}
type DashboardReportLine struct {
	Id  string // report id
	URL string // link to report details
}

func handleGetDashboard(templatesPath string, a *simba.Agent, rlr ReportListingRepository) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "dashboard.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
		var result DashboardPage
		rpts, err := rlr.AllReports(a.UserReportsFilter(user.Id))
		if err != nil {
			log.Printf("%s: %s: allReports: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		for _, rpt := range rpts {
			tl := DashboardTurnLine{
				Turn: rpt.Turn,
				Reports: []DashboardReportLine{
					DashboardReportLine{
						Id:  rpt.Id,
						URL: fmt.Sprintf("/reports/%s", rpt.Id),
					},
				},
			}
			result.Reports = append(result.Reports, tl)
		}
		sort.Slice(result.Reports, func(i, j int) bool {
			return result.Reports[i].Turn > result.Reports[j].Turn
		})

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
	}
}

func handleGetHero02(templatesPath string, a *simba.Agent) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "hero02.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		if r.URL.Path != "/" {
			log.Printf("%s: %s: get /... hack\n", r.Method, r.URL.Path)
			// this is stupid, but Go treats "GET /" as a wild-card not-found match.
			// we already have a handler for static files, so we'll just return a 404.
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		user, ok := a.CurrentUser(r)
		if ok && user.IsAuthenticated {
			log.Printf("%s: %s: user %q: ok && authenticate\n", r.Method, r.URL.Path, user.Id)
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		log.Printf("%s: %s: !(ok && authenticated)\n", r.Method, r.URL.Path)

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

type ReportListingPage struct {
	Page struct {
		Title string
	}
	ReportList reports.Listing
}

type ReportListingRepository interface {
	AllReports(authorize func(reports.Report) bool) (reports.Listing, error)
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

// returns a handler that will serve a static file if one exists, otherwise return not found.
func handleStaticFiles(prefix, root string, debug bool) http.HandlerFunc {
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

type TurnsListingPage struct {
	Page struct {
		Title string
	}
	Turns turns.Listing
}

type TurnsListingRepository interface {
	AllTurns(authorize func(turn turns.Turn) bool) (turns.Listing, error)
}

func handleTurnsListing(templatesPath string, a *simba.Agent, tlr TurnsListingRepository) http.Handler {
	templateFiles := []string{
		filepath.Join(templatesPath, "turns_listing.gohtml"),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		user, ok := a.CurrentUser(r)
		if !ok {
			log.Printf("%s: %s: currentUser: not ok\n", r.Method, r.URL.Path)
			http.Redirect(w, r, "/", http.StatusFound)
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

		var result TurnsListingPage
		result.Turns, err = tlr.AllTurns(a.UserTurnsFilter(user.Id))
		sort.Slice(result.Turns, func(i, j int) bool {
			return result.Turns[i].Id > result.Turns[j].Id
		})

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
