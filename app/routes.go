// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package app

import "net/http"

// todo: implement https://go.dev/blog/routing-enhancements

func (a *App) Routes() *http.ServeMux {
	mux := http.NewServeMux() // default mux, no routes

	mux.HandleFunc("GET /", a.getHero("02"))
	mux.HandleFunc("GET /dashboard", handleGetDashboard(a.paths.templates, a.policies, a.stores.reports))
	mux.HandleFunc("GET /features", a.getFeatures())
	mux.HandleFunc("GET /login", a.getLogin())
	mux.HandleFunc("POST /login", a.postLogin())
	mux.HandleFunc("GET /logout", a.getLogout())
	mux.HandleFunc("POST /logout", a.postLogout())
	mux.Handle("GET /reports", handleReportsListing(a.paths.templates, a.policies, a.stores.reports))
	mux.Handle("GET /reports/0991", handleReportsListing(a.paths.templates, a.policies, a.stores.reports))

	mux.HandleFunc("GET /api/version", a.handleVersion())
	mux.HandleFunc("GET /api/login/{name}/{secret}", a.apiGetLogin())

	return mux
}
