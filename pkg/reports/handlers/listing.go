// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package reports

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	domain "github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/simba"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
)

type ListingPage struct {
	Page struct {
		Title string
	}
	Reports []Listing
}

type Listing struct {
	Id  string
	URL string
}

func HandleGetReportsListing(templatesPath string, a *simba.Agent, repo interface {
	AllClanReports(cid string) ([]domain.Report, error)
}) http.Handler {
	templateFiles := []string{
		filepath.Join(templatesPath, "reports.gohtml"),
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		user, ok := a.CurrentUser(r)
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

		// the logic for the "service" should be a bunch of simple calls to the repository.
		allClanReports, err := repo.AllClanReports(user.Clan)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("%s: %s: clan %q: allClanReports: %v", r.Method, r.URL.Path, user.Clan, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			log.Printf("%s: %s: clan %q: allReports: no rows", r.Method, r.URL.Path, user.Clan)
		}
		sort.Slice(allClanReports, func(i, j int) bool {
			return allClanReports[i].Id > allClanReports[j].Id
		})

		var payload ListingPage
		for _, rpt := range allClanReports {
			payload.Reports = append(payload.Reports, Listing{
				Id:  rpt.Id,
				URL: fmt.Sprintf("/reports/%s", rpt.Id),
			})
		}

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
	})
}
