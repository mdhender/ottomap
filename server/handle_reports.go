// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"bytes"
	"github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/simba"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

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
