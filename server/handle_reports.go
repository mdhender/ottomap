// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"bytes"
	"github.com/mdhender/ottomap/domains/rbac"
	"github.com/mdhender/ottomap/sessions"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
)

type ReportListingPage struct {
	Page struct {
		Title string
	}
	ReportList ReportListing
}

func handleReportsListing(templatesPath string, rlr ReportListingRepository) http.Handler {
	templateFiles := []string{
		filepath.Join(templatesPath, "reports.gohtml"),
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		user := sessions.User(r.Context())
		if !user.IsAuthenticated {
			log.Printf("%s: %s: user is not authenticated (missing middleware?)\n", r.Method, r.URL.Path)
			//http.Redirect(w, r, "/logout", http.StatusFound)
			//return
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
		result.ReportList, err = rlr.AllReports(user.Roles)

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

type ReportListingRepository interface {
	AllReports(roles rbac.Roles) (ReportListing, error)
}

// ReportListing is a listing of reports that a User is allowed to view.
type ReportListing []Report

func (rl ReportListing) Len() int {
	return len(rl)
}

func (rl ReportListing) Less(i, j int) bool {
	return rl[i].Less(rl[j])
}

func (rl ReportListing) Swap(i, j int) {
	rl[i], rl[j] = rl[j], rl[i]
}

// Report is the metadata for a report.
type Report struct {
	Id     string // report id (e.g. 0991-02.0991)
	Turn   string // turn id formatted as YYY-MM (e.g. 901-02)
	Clan   string // clan id (e.g. 0991)
	Status string // status of report (e.g. "pending")
}

func (r Report) Less(other Report) bool {
	return r.Id < other.Id
}

// aqua is a mock implementation of ReportListingRepository.
type aquaReportListingRepository struct {
	Reports []Report
}

func NewAquaReportListingRepository() ReportListingRepository {
	return &aquaReportListingRepository{
		Reports: []Report{
			{"900-06.0991", "900-06", "0991", "Pending"},
			{"900-05.0991", "900-05", "0991", "Complete"},
			{"900-04.0991", "900-04", "0991", "Complete"},
			{"900-03.0991", "900-03", "0991", "Complete"},
			{"900-02.0991", "900-02", "0991", "Complete"},
			{"900-01.0991", "900-01", "0991", "Complete"},
			{"899-12.0991", "899-12", "0991", "Complete"},
		},
	}
}

func (a *aquaReportListingRepository) AllReports(roles rbac.Roles) (ReportListing, error) {
	var rl ReportListing
	for _, rpt := range a.Reports {
		rl = append(rl, rpt)
	}
	sort.Sort(rl)
	return rl, nil
}
