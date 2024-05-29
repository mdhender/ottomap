// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sessions

import (
	"context"
	"github.com/mdhender/ottomap/users"
)

// userContextKey is the context key type for storing User in context.Context.
type userContextKey string

// User returns the current user from the request context.
func User(ctx context.Context) users.User {
	user, ok := ctx.Value(userContextKey("user")).(users.User)
	if !ok {
		return users.User{}
	}
	return user
}

// AddUser adds a User to the request context.
func (sm *Manager) AddUser(ctx context.Context, user users.User) context.Context {
	return context.WithValue(ctx, userContextKey("user"), user)
}

// GetUser returns the current user from the request context.
func (sm *Manager) GetUser(ctx context.Context) (users.User, bool) {
	user, ok := ctx.Value(userContextKey("user")).(users.User)
	if !ok {
		return users.User{}, false
	}
	return user, true
}
