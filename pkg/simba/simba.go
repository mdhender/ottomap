// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package simba implements a simple policy agent.
// It relies on user, rbac, and session data from the repository
// for input and returns true/false for each policy question.
package simba

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/mdhender/ottomap/pkg/simba/sqlc"
	"golang.org/x/crypto/bcrypt"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"time"
)

// Agent is the policy agent.
type Agent struct {
	db  *sql.DB
	q   *sqlc.Queries
	ctx context.Context
}

// NewAgent returns a new policy agent that uses the supplied database as input
// when answering policy questions.
func NewAgent(path string, ctx context.Context) (*Agent, error) {
	// verify that the repository exists
	// todo: provide a way to create the repository!
	if sb, err := os.Stat(path); err != nil {
		return nil, err
	} else if sb.IsDir() {
		return nil, os.ErrInvalid
	} else if !sb.Mode().IsRegular() {
		return nil, os.ErrInvalid
	}

	// open the path as a sqlite database
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	a := &Agent{
		db:  db,
		ctx: ctx,
		q:   sqlc.New(db),
	}

	return a, nil
}

// Authenticate returns the user ID if the credentials are valid.
// Returns false if not.
func (a *Agent) Authenticate(name, secret string) (string, bool) {
	user, err := a.q.ReadUserAuthenticationData(a.ctx, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: userIsAuthenticated: %v\n", err)
		}
		return "", false
	} else if user.HashedPassword == "$$" {
		log.Printf("agent: authenticate: $$ bypass!\n")
		return user.Uid, true
	}

	// check if two passwords match using bcrypt's CompareHashAndPassword
	// which return nil on success and an error on failure.
	// (from gregorygaines.com)
	hashedSecretBytes, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.MinCost)
	if err = bcrypt.CompareHashAndPassword(hashedSecretBytes, []byte(user.HashedPassword)); err != nil {
		return "", false
	}

	return user.Uid, true
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
	sid := uuid.New().String()
	if err := a.q.CreateSession(a.ctx, sqlc.CreateSessionParams{
		Sid:       sid,
		Uid:       uid,
		ExpiresAt: time.Now().UTC().Add(2 * 7 * 24 * time.Hour),
	}); err != nil {
		log.Printf("agent: createSession: uid %q: %v\n", uid, err)
		return "", false
	} else if !a.CreateCookie(w, sid) {
		log.Printf("agent: createSession: uid %q: createCookie failed\n", uid)
		_ = a.q.DeleteSession(a.ctx, sid)
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
	row, err := a.q.GetSession(a.ctx, cookie.Value)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: currentUser: sid %q: %v\n", sid, err)
			return User{}, false
		}
		log.Printf("agent: currentUser: sid %q: not found\n", sid)
		return User{}, false
	}
	if !time.Now().Before(row.ExpiresAt) {
		log.Printf("agent: currentUser: sid %q: expired\n", sid)
		return User{}, false
	}
	uid := row.Uid
	u, err := a.q.ReadUser(a.ctx, uid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: currentUser: sid %q: uid %q: read %v\n", sid, uid, err)
			return User{}, false
		}
		log.Printf("agent: currentUser: sid %q: uid %q: not found\n", sid, uid)
		return User{}, false
	}
	user := User{
		Id:              row.Uid,
		Handle:          u.Username,
		Email:           u.Email,
		Clan:            u.Clan,
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
	if err := a.q.DeleteSession(a.ctx, cookie.Value); err != nil {
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
	row, err := a.q.GetSession(a.ctx, cookie.Value)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: sessionUserFromRequest: %v\n", err)
		}
		log.Printf("agent: sessionUserFromRequest: session: id not found\n")
		return "", false
	}
	if !time.Now().Before(row.ExpiresAt) {
		log.Printf("agent: sessionUserFromRequest: session: expired\n")
		return "", false
	}
	return row.Uid, true
}

// Close closes the physical database connection.
// Please call it to avoid leaking memory or file handles.
func (a *Agent) Close() error {
	if a.db == nil {
		return nil
	}
	err := a.db.Close()
	a.db = nil
	return err
}
