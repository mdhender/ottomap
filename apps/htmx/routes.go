// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) Routes() (*http.ServeMux, error) {
	mux := http.NewServeMux() // default mux, no routes

	mux.HandleFunc("GET /", getHomePage(a.paths.templates, "", a.paths.assets, false))

	return mux, nil
}

func getHomePage(templatesPath string, prefix, root string, debug bool) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "home_page.gohtml"),
	}
	_ = templateFiles

	return func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		// this is stupid, but Go treats "GET /" as a wild-card not-found match.
		if r.URL.Path != "/" {
			file := filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))
			if debug {
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
	}
}
