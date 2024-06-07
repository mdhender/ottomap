// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package spa

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type SPA struct {
	paths struct {
		public string // path to public files
	}
}

func New(public string) (*SPA, error) {
	e := &SPA{}
	e.paths.public = public

	if len(e.paths.public) == 0 {
		return nil, fmt.Errorf("missing path to public files")
	} else if e.paths.public != strings.TrimSpace(e.paths.public) {
		return nil, fmt.Errorf("path to public files must not contain leading or trailing spaces")
	} else if path, err := filepath.Abs(e.paths.public); err != nil {
		return nil, err
	} else if sb, err := os.Stat(path); err != nil {
		return nil, err
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("path to public files is not a directory")
	} else {
		e.paths.public = path
	}

	return e, nil
}

func (e *SPA) Routes() (*http.ServeMux, error) {
	mux := http.NewServeMux() // default mux, no routes

	mux.HandleFunc("GET /", getDistFiles("", e.paths.public))

	return mux, nil
}

// returns a handler that will serve a file if it exists, otherwise serve index.html.
func getDistFiles(prefix, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		var file string
		if r.URL.Path == "/" {
			file = filepath.Join(root, "index.html")
		} else {
			file = filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))
		}
		log.Printf("%s: %s: %s\n", r.Method, r.URL.Path, file)

		stat, err := os.Stat(file)
		if err != nil {
			file = filepath.Join(root, "index.html")
			stat, err = os.Stat(file)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
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
