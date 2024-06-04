// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"github.com/mdhender/ottomap/pkg/simba"
	"log"
	"net/http"
)

func HandleGetDetail(templatesPath string, a *simba.Agent, repo Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		turnId := r.PathValue("turnId")
		log.Printf("%s: %s: id %q\n", r.Method, r.URL.Path, turnId)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("to be implemented"))
	})
}
