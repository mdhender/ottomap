// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"bytes"
	"github.com/mdhender/ottomap/pkg/simba"
	turns "github.com/mdhender/ottomap/pkg/turns/domain"
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
	Turns []Listing
}

type Listing struct {
	Id    string // turn id (e.g. 0991-02)
	Turn  string // display value for turn id formatted as YYY-MM (e.g. 901-02)
	Year  int    // year of turn (e.g. 901)
	Month int    // month of turn (e.g. 02)
	URL   string // url to turn (e.g. /turns/0901-02)
}

type Repository interface {
	AllTurns(authorize func(turn turns.Turn) bool) ([]turns.Turn, error)
}

func HandleGetListing(templatesPath string, a *simba.Agent, repo Repository) http.Handler {
	templateFiles := []string{
		filepath.Join(templatesPath, "turns_listing.gohtml"),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		user, ok := a.CurrentUser(r)
		if !(ok && user.IsAuthenticated) {
			log.Printf("%s: %s: user: not authenticated\n", r.Method, r.URL.Path)
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

		allTurns, err := repo.AllTurns(a.UserTurnsFilter(user.Id))
		if err != nil {
			log.Printf("%s: %s: turns listing: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		sort.Slice(allTurns, func(i, j int) bool {
			return allTurns[i].Id > allTurns[j].Id
		})

		var result ListingPage
		for _, turn := range allTurns {
			result.Turns = append(result.Turns, Listing{
				Id:    turn.Id,
				Turn:  turn.Turn,
				Year:  turn.Year,
				Month: turn.Month,
				URL:   turn.URL,
			})
		}

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
