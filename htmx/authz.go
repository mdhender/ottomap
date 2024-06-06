// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import "net/http"

type AuthRepo interface {
	CurrentUser(*http.Request) AuthUser
}

type AuthUser struct{}

type AuthStore struct{}

func (a *AuthStore) CurrentUser(r *http.Request) AuthUser {
	return AuthUser{}
}

func (a AuthUser) IsAuthenticated() bool {
	return true
}

func (a AuthUser) ClanId() string {
	return "0138"
}
