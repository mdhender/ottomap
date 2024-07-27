// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"bytes"
	"github.com/mdhender/ottomap/stores/ffs"
	tmpls "github.com/mdhender/ottomap/templates/htmx"
	"github.com/mdhender/ottomap/templates/tw"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) Routes() (*http.ServeMux, error) {
	mux := http.NewServeMux() // default mux, no routes

	mux.HandleFunc("GET /", getHomePage(a.paths.templates, "", a.paths.assets, true, true))
	mux.HandleFunc("GET /login/{clan}/{id}", getLogin(a.sessions, true))
	mux.HandleFunc("GET /logout", getLogout())

	// https://datatracker.ietf.org/doc/html/rfc9110 for POST vs PUT

	mux.HandleFunc("GET /tn3", authonly(a.sessions, getTurnListing(a.paths.templates, a.store, a.sessions)))

	mux.HandleFunc("GET /tn3/{turnId}", authonly(a.sessions, getTurnDetails(a.paths.templates, a.store, a.sessions)))
	mux.HandleFunc("DELETE /tn3/{turnId}", authonly(a.sessions, handleNotImplemented()))

	mux.HandleFunc("GET /tn3/{turnId}/{clanId}", authonly(a.sessions, handleNotImplemented()))
	mux.HandleFunc("DELETE /tn3/{turnId}/{clanId}", authonly(a.sessions, handleNotImplemented()))

	mux.HandleFunc("GET /tn3/{turnId}/{clanId}/map", authonly(a.sessions, handleNotImplemented()))
	mux.HandleFunc("POST /tn3/{turnId}/{clanId}/map", authonly(a.sessions, handleNotImplemented()))
	mux.HandleFunc("PUT /tn3/{turnId}/{clanId}/map", authonly(a.sessions, handleNotImplemented()))
	mux.HandleFunc("DELETE /tn3/{turnId}/{clanId}/map", authonly(a.sessions, handleNotImplemented()))

	mux.HandleFunc("GET /tn3/{turnId}/{clanId}/report", authonly(a.sessions, handleNotImplemented()))
	mux.HandleFunc("POST /tn3/{turnId}/{clanId}/report", authonly(a.sessions, handleNotImplemented()))
	mux.HandleFunc("PUT /tn3/{turnId}/{clanId}/report", authonly(a.sessions, handleNotImplemented()))
	mux.HandleFunc("DELETE /tn3/{turnId}/{clanId}/report", authonly(a.sessions, handleNotImplemented()))

	return mux, nil
}

func handleNotImplemented() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: not implemented\n", r.Method, r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

// the home page is displayed when the URL is empty or doesn't match any other route.

func getHomePage(templatesPath string, prefix, root string, debug, debugAssets bool) http.HandlerFunc {
	// we can display two types of content. the first is for unauthenticated users,
	// the second is for authenticated users.

	anonTemplateFiles := []string{
		filepath.Join(templatesPath, "layout.gohtml"),
	}
	authTemplateFiles := []string{
		filepath.Join(templatesPath, "layout.gohtml"),
	}
	_, _ = anonTemplateFiles, authTemplateFiles

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		// this is stupid, but Go treats "GET /" as a wild-card not-found match.
		if r.URL.Path != "/" {
			file := filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))
			if debugAssets {
				log.Printf("%s: %s: assets\n", r.Method, r.URL.Path)
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
			return
		}

		// Parse the template file
		tmpl, err := template.ParseFiles(anonTemplateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var payload tmpls.Layout
		payload.Banner = tmpls.Banner{
			Title: "Ottomap",
			Slug:  "trying to map the right thing...",
		}
		payload.MainMenu = tmpls.MainMenu{
			Items: []tmpls.MenuItem{
				{Label: "Main pages", Link: "#",
					Children: []tmpls.MenuItem{
						{Label: "Blog", Link: "#"},
						{Label: "Archives", Link: "#"},
						{Label: "Categories", Link: "#"},
					},
				},
				{Label: "Blog topics", Link: "#",
					Children: []tmpls.MenuItem{
						{Label: "Web design", Link: "#"},
						{Label: "Accessibility", Link: "#"},
						{Label: "CMS solutions", Link: "#"},
					},
				},
				{Label: "Extras", Link: "#",
					Children: []tmpls.MenuItem{
						{Label: "Music archive", Link: "#"},
						{Label: "Photo gallery", Link: "#"},
						{Label: "Poems and lyrics", Link: "#"},
					},
				},
				{Label: "Community", Link: "#",
					Children: []tmpls.MenuItem{
						{Label: "Guestbook", Link: "#"},
						{Label: "Members", Link: "#"},
						{Label: "Link collections", Link: "#"},
					},
				},
			},
			Releases: tmpls.Releases{
				DT: tmpls.Link{Label: "Releases", Link: "https://github.com/mdhender/ottomap/releases", Target: "_blank"},
				DDs: []tmpls.Link{
					{Label: "v0.13.8", Link: "https://github.com/mdhender/ottomap/releases/tag/v0.13.8", Target: "_blank"},
				},
			},
		}
		payload.Sidebar.LeftMenu = tmpls.LeftMenu{
			Items: []tmpls.MenuItem{
				{Label: "Left menu", Class: "sidemenu",
					Children: []tmpls.MenuItem{
						{Label: "First page", Link: "#"},
						{Label: "Second page", Link: "#"},
						{Label: "Third page with subs", Link: "#",
							Children: []tmpls.MenuItem{
								{Label: "First subpage", Link: "#"},
								{Label: "Second subpage", Link: "#"},
							}},
						{Label: "Fourth page", Link: "#"},
					},
				},
			},
		}
		payload.Sidebar.RightMenu = tmpls.RightMenu{
			Items: []tmpls.MenuItem{
				{Label: "Right menu", Class: "sidemenu",
					Children: []tmpls.MenuItem{
						{Label: "Sixth page", Link: "#"},
						{Label: "Seventh page", Link: "#"},
						{Label: "Another page", Link: "#"},
						{Label: "The last one", Link: "#"},
					},
				},
				{Label: "Sample links",
					Children: []tmpls.MenuItem{
						{Label: "Sample link 1", Link: "#"},
						{Label: "Sample link 2", Link: "#"},
						{Label: "Sample link 3", Link: "#"},
						{Label: "Sample link 4", Link: "#"},
					},
				},
			},
		}
		payload.Sidebar.Notice = &tmpls.Notice{
			Label: "Account",
			Lines: []string{
				"You aren't logged in. Please use your secret link to log in.",
				"If you don't have an account, please visit the Discord server to request an account.",
			},
		}
		payload.Footer = tmpls.Footer{
			Author:        "Michael D Henderson",
			CopyrightYear: "2024",
		}

		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
		buf := &bytes.Buffer{}

		// execute the template with our payload
		err = tmpl.ExecuteTemplate(buf, "layout", payload)
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

func getLogin(sessionManager *sessionManager_t, debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		id := r.PathValue("id")
		if id == "" {
			// delete any cookies that might be set.
			http.SetCookie(w, &http.Cookie{
				Path:    "/",
				Name:    "ottomap",
				Expires: time.Unix(0, 0),
			})
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		sess := sessionManager.getSession(id)
		if !sess.isValid() {
			// delete any cookies that might be set.
			http.SetCookie(w, &http.Cookie{
				Path:    "/",
				Name:    "ottomap",
				Expires: time.Unix(0, 0),
			})
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// set a new cookie with the new expiration date
		http.SetCookie(w, &http.Cookie{
			Path:    "/",
			Name:    "ottomap",
			Value:   sess.id,
			Expires: sess.expires,
		})

		http.Redirect(w, r, "/index.html", http.StatusSeeOther)
	}
}

func getLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		// delete any cookies that might be set.
		http.SetCookie(w, &http.Cookie{
			Path:    "/",
			Name:    "ottomap",
			Expires: time.Unix(0, 0),
		})

		http.Redirect(w, r, "/index.html", http.StatusSeeOther)
	}
}

type SessionManager_i interface {
	currentUser(r *http.Request) session_t
}

type AllTurns_i interface {
	GetTurnListing(id string) ([]ffs.Turn_t, error)
	GetTurnDetails(id string, turnId string) (ffs.TurnDetail_t, error)
}

func getTurnListing(templatesPath string, s AllTurns_i, sm SessionManager_i) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "layout.gohtml"),
		filepath.Join(templatesPath, "turn_list.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		log.Printf("%s: %s: session: id %q\n", r.Method, r.URL.Path, sm.currentUser(r).id)
		if !sm.currentUser(r).isAuthenticated() {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		turns, _ := s.GetTurnListing(sm.currentUser(r).id)
		log.Printf("%s: %s: turns %+v\n", r.Method, r.URL.Path, turns)

		var payload tw.Layout_t

		var content tw.TurnList_t
		for _, turn := range turns {
			content.Turns = append(content.Turns, turn.Id)
		}
		payload.Content = content

		// Parse the template file
		tmpl, err := template.ParseFiles(templateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
		buf := &bytes.Buffer{}

		// execute the template with our payload
		err = tmpl.ExecuteTemplate(buf, "layout", payload)
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

func getTurnDetails(templatesPath string, s AllTurns_i, sm SessionManager_i) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "layout.gohtml"),
		filepath.Join(templatesPath, "turn_details.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		log.Printf("%s: %s: session: id %q\n", r.Method, r.URL.Path, sm.currentUser(r).id)
		if !sm.currentUser(r).isAuthenticated() {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		turnId := r.PathValue("turnId")
		turn, _ := s.GetTurnDetails(sm.currentUser(r).id, turnId)
		log.Printf("%s: %s: turns %+v\n", r.Method, r.URL.Path, turn)

		var payload tw.Layout_t
		var content tw.TurnDetails_t
		content.Id = turnId
		for _, clan := range turn.Clans {
			content.Clans = append(content.Clans, clan)
		}
		payload.Content = content

		// Parse the template file
		tmpl, err := template.ParseFiles(templateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
		buf := &bytes.Buffer{}

		// execute the template with our payload
		err = tmpl.ExecuteTemplate(buf, "layout", payload)
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
