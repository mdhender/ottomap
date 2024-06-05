// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package clans

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	reports "github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/simba"
	turns "github.com/mdhender/ottomap/pkg/turns/domain"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
)

type DetailPage struct {
	Page struct {
		Title string
	}
	NoTurns   bool
	NoReports bool
	Clan      string
	URL       string
	Turns     []ClanTurnDetail
}

type ClanTurnDetail struct {
	Id      string
	Turn    string
	Reports []ClanReportDetail
	URL     string
}

type ClanReportDetail struct {
	Id        string // report id
	Separator string
	URL       string // link to report details
}

func HandleGetClanDetail(templatesPath string, a *simba.Agent, repo interface {
	AllClanReports(cid string) ([]reports.Report, error)
	AllTurns() ([]turns.Turn, error)
}) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "clan_detail.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		user, ok := a.CurrentUser(r)
		if !(ok && user.IsAuthenticated) {
			log.Printf("%s: %s: currentUser: ok %v: authenticated %v\n", r.Method, r.URL.Path, ok, user.IsAuthenticated)
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

		allTurns, err := repo.AllTurns()
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("%s: %s: clan %q: allTurns: %v", r.Method, r.URL.Path, user.Clan, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			log.Printf("%s: %s: clan %q: allTurns: no rows", r.Method, r.URL.Path, user.Clan)
		}
		sort.Slice(allTurns, func(i, j int) bool {
			return allTurns[i].Id > allTurns[j].Id
		})

		allClanReports, err := repo.AllClanReports(user.Clan)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("%s: %s: clan %q: allClanReports: %v", r.Method, r.URL.Path, user.Clan, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			log.Printf("%s: %s: clan %q: allClanReports: no rows", r.Method, r.URL.Path, user.Clan)
		}
		sort.Slice(allClanReports, func(i, j int) bool {
			return allClanReports[i].Id > allClanReports[j].Id
		})

		var payload DetailPage
		payload.Page.Title = fmt.Sprintf("Clan: %s", user.Clan)
		payload.Clan = user.Clan

		for _, turn := range allTurns {
			ctd := ClanTurnDetail{
				Id:      turn.Id,
				Turn:    turn.Turn,
				Reports: []ClanReportDetail{},
				URL:     fmt.Sprintf("/turns/%s", turn.Id),
			}
			payload.Turns = append(payload.Turns, ctd)
		}
		payload.NoTurns = len(allTurns) == 0

		for n, rpt := range allClanReports {
			crd := ClanReportDetail{
				Id:  rpt.Id,
				URL: fmt.Sprintf("/reports/%s", rpt.Id),
			}
			if n > 0 {
				crd.Separator = ", "
			}
			for _, turn := range payload.Turns {
				if turn.Id == rpt.Turn {
					turn.Reports = append(turn.Reports, crd)
					break
				}
			}
		}
		payload.NoReports = len(allClanReports) == 0

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
