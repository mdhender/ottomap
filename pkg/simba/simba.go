// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package simba implements a simple policy agent.
// It relies on user, rbac, and session data from the repository
// for input and returns true/false for each policy question.
package simba

import (
	"context"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"time"
)

// Agent is the policy agent.
type Agent struct {
	db *sqlc.DB
}

// NewAgent returns a new policy agent that uses the supplied database as input
// when answering policy questions.
func NewAgent(db *sqlc.DB, ctx context.Context) (*Agent, error) {
	a := &Agent{
		db: db,
	}

	return a, nil
}

// Authenticate returns the user ID if the credentials are valid.
// Returns false if not.
func (a *Agent) Authenticate(handle, plaintextSecret string) (uid string, ok bool) {
	uid, err := a.db.AuthenticateUserHandle(handle, plaintextSecret)
	if err != nil {
		log.Printf("agent: authenticate: %v\n", err)
		return "", false
	}
	return uid, true
}

// CreateCookie is not really policy, either.
func (a *Agent) CreateCookie(w http.ResponseWriter, sid string) bool {
	http.SetCookie(w, &http.Cookie{
		Name:     "simba-session",
		Path:     "/",
		Value:    sid,
		Expires:  time.Now().UTC().Add(2 * 7 * 24 * time.Hour),
		HttpOnly: true, // Set the HttpOnly flag to prevent client-side script access
		Secure:   true, // Set the Secure flag to ensure the cookie is only sent over HTTPS
		SameSite: http.SameSiteLaxMode,
	})
	return true
}

// CreateSession is not really policy, either.
// Side effect of creating a session is setting a cookie.
func (a *Agent) CreateSession(w http.ResponseWriter, uid string) (string, bool) {
	sid, err := a.db.CreateSession(uid)
	if err != nil {
		log.Printf("agent: createSession: uid %q: %v\n", uid, err)
		return "", false
	} else if !a.CreateCookie(w, sid) {
		log.Printf("agent: createSession: uid %q: createCookie failed\n", uid)
		_ = a.db.DeleteSession(sid)
		return "", false
	}
	return sid, true
}

// CurrentUser is not really policy, either.
func (a *Agent) CurrentUser(r *http.Request) (User, bool) {
	cookie, err := r.Cookie("simba-session")
	if err != nil {
		log.Printf("agent: currentUser: no cookie\n")
		return User{}, false
	}
	sid := cookie.Value
	log.Printf("agent: currentUser: sid %q\n", sid)
	uid, exp, err := a.db.ReadSession(cookie.Value)
	if err != nil {
		log.Printf("agent: currentUser: sid %q: %v\n", sid, err)
		return User{}, false
	} else if uid == "" {
		log.Printf("agent: currentUser: sid %q: not found\n", sid)
		return User{}, false
	} else if !time.Now().Before(exp) {
		log.Printf("agent: currentUser: sid %q: expired\n", sid)
		return User{}, false
	}
	handle, email, err := a.db.ReadUser(uid)
	if err != nil {
		log.Printf("agent: currentUser: sid %q: uid %q: user: %v\n", sid, uid, err)
		return User{}, false
	} else if handle == "" {
		log.Printf("agent: currentUser: sid %q: uid %q: user: not found\n", sid, uid)
		return User{}, false
	}
	clan, err := a.db.ReadUserClan(uid)
	if err != nil {
		log.Printf("agent: currentUser: sid %q: uid %q: clan: %v\n", sid, uid, err)
		return User{}, false
	} else if clan == "" {
		log.Printf("agent: currentUser: sid %q: uid %q: clan: not found\n", sid, uid)
		return User{}, false
	}
	user := User{
		Id:              uid,
		Handle:          handle,
		Email:           email,
		Clan:            clan,
		IsAuthenticated: true,
	}
	log.Printf("agent: currentUser: sid %q: uid %q: found\n", sid, uid)
	return user, true
}

// DeleteCookie is not really policy, either.
func (a *Agent) DeleteCookie(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:   "simba-session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
}

// DeleteSession is not really policy, either.
func (a *Agent) DeleteSession(r *http.Request) {
	cookie, err := r.Cookie("simba-session")
	if err != nil || cookie.Value == "" {
		return
	}
	if err := a.db.DeleteSession(cookie.Value); err != nil {
		log.Printf("agent: deleteSession: %v\n", err)
	}
}

// SessionUserFromRequest is not really policy, either.
func (a *Agent) SessionUserFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("simba-session")
	if err != nil {
		log.Printf("agent: sessionUserFromRequest: no cookie\n")
		return "", false
	}
	uid, exp, err := a.db.ReadSession(cookie.Value)
	if err != nil {
		log.Printf("agent: sessionUserFromRequest: %v\n", err)
		return "", false
	} else if uid == "" {
		log.Printf("agent: sessionUserFromRequest: session: id not found\n")
		return "", false
	} else if !time.Now().Before(exp) {
		log.Printf("agent: sessionUserFromRequest: session: expired\n")
		return "", false
	}
	return uid, true
}
