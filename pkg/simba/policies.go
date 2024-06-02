// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package simba

import (
	"database/sql"
	"errors"
	"github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/simba/sqlc"
	"log"
	"net/http"
	"strings"
	"time"
)

// RequestIsAuthenticated returns true if the request contains is a valid session that has not expired.
func (a *Agent) RequestIsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("simba-session")
	if err != nil {
		log.Printf("agent: requestIsAuthenticated: no cookie\n")
		return false
	}
	return a.SessionIsAuthenticated(cookie.Value)
}

// SessionIsAuthenticated returns true if the session is valid and has not expired.
func (a *Agent) SessionIsAuthenticated(sid string) bool {
	row, err := a.q.GetSession(a.ctx, sid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: sessionIsAuthenticated: %v\n", err)
		}
		log.Printf("agent: sessionIsAuthenticated: session: id not found\n")
		return false
	}
	if !time.Now().Before(row.ExpiresAt) {
		log.Printf("agent: sessionIsAuthenticated: session: expired\n")
		return false
	}
	return true
}

func (a *Agent) UserCanViewReport(uid string, rpt reports.Report) bool {
	user, err := a.q.ReadUser(a.ctx, uid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: userIsAuthenticated: %v\n", err)
		}
		return false
	}
	if user.Clan == rpt.Clan {
		return true
	}
	clans, err := a.q.ReadUserRole(a.ctx, sqlc.ReadUserRoleParams{
		Uid: uid,
		Rid: "clans",
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: userIsAuthenticated: %v\n", err)
		}
		return false
	}
	for _, clan := range strings.Split(clans, ",") {
		if clan == rpt.Clan {
			return true
		}
	}
	return a.UserIsAdministrator(uid)
}

// UserIsAdministrator returns true if the user is an administrator.
func (a *Agent) UserIsAdministrator(uid string) bool {
	role, err := a.q.ReadUserRole(a.ctx, sqlc.ReadUserRoleParams{
		Uid: uid,
		Rid: "administrator",
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: userIsAuthenticated: %v\n", err)
		}
		return false
	}
	return role == "true"
}

func (a *Agent) UserReportsFilter(uid string) func(reports.Report) bool {
	user, err := a.q.ReadUser(a.ctx, uid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: userIsAuthenticated: %v\n", err)
		}
		return func(reports.Report) bool { return false }
	}

	role, err := a.q.ReadUserRole(a.ctx, sqlc.ReadUserRoleParams{
		Uid: uid,
		Rid: "clans",
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("agent: userReportsFilter: %v\n", err)
		}
		return func(reports.Report) bool { return false }
	}

	clans := strings.Split(role, ",")

	return func(rpt reports.Report) bool {
		if rpt.Clan == user.Clan {
			return true
		}
		for _, clan := range clans {
			if rpt.Clan == clan {
				return true
			}
		}
		return false
	}
}
