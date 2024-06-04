// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"bytes"
	"fmt"
	"github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/simba"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type TurnDetailPayload struct {
	Page struct {
		Title string
	}
	Title                    string
	Reports                  []TurnDetailReportDetail
	DeleteReportsButtonLabel string
}

type TurnDetailReportDetail struct {
	Id     string // turn id (e.g. 0991-02)
	Clan   string
	Status string
	Date   string // yyyy/mm/dd
	URL    string // url to turn (e.g. /turns/0901-02)
}

func HandleGetDetail(templatesPath string, a *simba.Agent, repo interface {
	UserTurnReports(userId, turnId string) ([]reports.Report, error)
}) http.Handler {
	templateFiles := []string{
		filepath.Join(templatesPath, "turn_detail.gohtml"),
	}

	// todo: set this when we're running in production
	var doOnceTemplate *template.Template

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		user, ok := a.CurrentUser(r)
		if !(ok && user.IsAuthenticated) {
			log.Printf("%s: %s: user: not authenticated\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// fetch path parameters
		turnId := r.PathValue("turnId")
		log.Printf("%s: %s: id %q\n", r.Method, r.URL.Path, turnId)

		// create the payload, reporting errors
		var payload TurnDetailPayload
		payload.Page.Title = fmt.Sprintf("Turn %s", turnId)
		payload.Title = fmt.Sprintf("Turn %s", turnId)
		var rpts []reports.Report
		if repo != nil {
			var err error
			if rpts, err = repo.UserTurnReports(user.Id, turnId); err != nil {
				log.Printf("%s: %s: userTurnReports: %v", r.Method, r.URL.Path, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		for _, rpt := range rpts {
			payload.Reports = append(payload.Reports, TurnDetailReportDetail{
				Id:     rpt.Id,
				Clan:   rpt.Clan,
				Status: rpt.Status,
				Date:   "2024/05/24",
				URL:    fmt.Sprintf("/reports/%s", rpt.Id),
			})
		}
		if len(payload.Reports) == 1 {
			payload.DeleteReportsButtonLabel = "Delete Report"
		} else if len(payload.Reports) > 1 {
			payload.DeleteReportsButtonLabel = "Delete Reports"
		}

		// parse the template file if needed, reporting errors
		var t *template.Template
		if doOnceTemplate == nil {
			var err error
			if t, err = template.ParseFiles(templateFiles...); err != nil {
				log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else {
			t = doOnceTemplate
		}

		// render the template using the payload and report all errors
		buf := &bytes.Buffer{}
		if err := t.Execute(buf, payload); err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	})
}
